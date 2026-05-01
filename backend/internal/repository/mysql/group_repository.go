package mysql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"mini_im/backend/internal/model"
)

var (
	ErrGroupNotFound            = errors.New("group not found")
	ErrGroupPermissionDenied    = errors.New("group permission denied")
	ErrGroupJoinRequestNotFound = errors.New("group join request not found")
	ErrGroupJoinRequestPending  = errors.New("group join request pending")
	ErrGroupAlreadyMember       = errors.New("already group member")
	ErrGroupFull                = errors.New("group is full")
	ErrGroupDissolved           = errors.New("group dissolved")
	ErrGroupMemberNotFound      = errors.New("group member not found")
	ErrGroupMemberMuted         = errors.New("group member muted")
	ErrGroupOwnerCannotLeave    = errors.New("group owner cannot leave")
	ErrDuplicateGroupNo         = errors.New("duplicate group no")
)

type GroupRepository struct {
	db *sql.DB
}

type GroupMessageContext struct {
	Group  model.Group
	Member model.GroupMember
	Sender model.GroupMessageSender
}

type GroupSettingsUpdate struct {
	AllowMemberInvite *bool
	MaxMembers        *int
}

type HandleJoinRequestResult struct {
	GroupID        int64
	ConversationID int64
	UserID         int64
	Status         string
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateGroupWithOwner(ctx context.Context, group *model.Group) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO conversations (conversation_id, conversation_type, ref_id, status)
VALUES (?, ?, ?, ?)
`, group.ConversationID, model.ConversationTypeGroup, group.GroupID, model.ConversationStatusNormal); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO `+"`groups`"+` (
  group_id, group_no, conversation_id, owner_id, name, avatar_url,
  max_members, allow_member_invite, status, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		group.GroupID,
		group.GroupNo,
		group.ConversationID,
		group.OwnerID,
		group.Name,
		group.AvatarURL,
		group.MaxMembers,
		group.AllowMemberInvite,
		group.Status,
		group.CreatedAt,
	); err != nil {
		if isDuplicateEntry(err) {
			return ErrDuplicateGroupNo
		}
		return err
	}

	if err := upsertGroupMember(ctx, tx, group.GroupID, group.OwnerID, model.GroupRoleOwner); err != nil {
		return err
	}
	if err := upsertConversationMemberWithRole(ctx, tx, group.ConversationID, group.OwnerID, model.ConversationMemberRoleOwner); err != nil {
		return err
	}
	if err := upsertConversationUserState(ctx, tx, group.ConversationID, group.OwnerID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *GroupRepository) Search(ctx context.Context, keyword string, userID int64, limit int) ([]model.Group, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []model.Group{}, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, group_id, group_no, conversation_id, owner_id, name, avatar_url,
  max_members, allow_member_invite, status, created_at, updated_at
FROM `+"`groups`"+`
WHERE status = ?
  AND (group_no = ? OR name LIKE ?)
ORDER BY CASE WHEN group_no = ? THEN 0 ELSE 1 END, created_at DESC
LIMIT ?
`, model.GroupStatusNormal, keyword, "%"+keyword+"%", keyword, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]model.Group, 0)
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepository) CreateJoinRequest(ctx context.Context, request *model.GroupJoinRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, request.GroupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}

	member, err := findGroupMemberForUpdate(ctx, tx, request.GroupID, request.UserID)
	if err != nil && !errors.Is(err, ErrGroupMemberNotFound) {
		return err
	}
	if member != nil && member.Status == model.GroupMemberStatusActive {
		return ErrGroupAlreadyMember
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO group_join_requests (request_id, group_id, user_id, message, status)
VALUES (?, ?, ?, ?, ?)
`, request.RequestID, request.GroupID, request.UserID, request.Message, model.GroupJoinRequestStatusPending)
	if err != nil {
		if isDuplicateEntry(err) {
			return ErrGroupJoinRequestPending
		}
		return err
	}
	return tx.Commit()
}

func (r *GroupRepository) ListJoinRequests(ctx context.Context, groupID int64, operatorID int64, limit int, offset int) ([]model.GroupJoinRequestWithUser, int64, error) {
	if err := r.requireRole(ctx, groupID, operatorID, model.GroupRoleOwner, model.GroupRoleAdmin); err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM group_join_requests
WHERE group_id = ?
`, groupID).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.GroupJoinRequestWithUser{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  gr.id, gr.request_id, gr.group_id, gr.user_id, gr.message, gr.status, gr.handled_by, gr.created_at, gr.updated_at,
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM group_join_requests gr
INNER JOIN users u ON u.user_id = gr.user_id AND u.deleted_at IS NULL
INNER JOIN user_profiles p ON p.user_id = u.user_id
WHERE gr.group_id = ?
ORDER BY gr.created_at DESC
LIMIT ? OFFSET ?
`, groupID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.GroupJoinRequestWithUser, 0)
	for rows.Next() {
		item, err := scanGroupJoinRequestWithUser(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GroupRepository) HandleJoinRequest(ctx context.Context, requestID int64, operatorID int64, accept bool) (*HandleJoinRequestResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	request, err := findJoinRequestForUpdate(ctx, tx, requestID)
	if err != nil {
		return nil, err
	}
	if request.Status != model.GroupJoinRequestStatusPending {
		return nil, ErrGroupJoinRequestNotFound
	}

	group, err := findGroupByIDForUpdate(ctx, tx, request.GroupID)
	if err != nil {
		return nil, err
	}
	if group.Status != model.GroupStatusNormal {
		return nil, ErrGroupDissolved
	}

	operator, err := findGroupMemberForUpdate(ctx, tx, request.GroupID, operatorID)
	if err != nil {
		return nil, ErrGroupPermissionDenied
	}
	if operator.Status != model.GroupMemberStatusActive || !isOwnerOrAdmin(operator.Role) {
		return nil, ErrGroupPermissionDenied
	}

	nextStatus := model.GroupJoinRequestStatusRejected
	if accept {
		nextStatus = model.GroupJoinRequestStatusAccepted
		member, err := findGroupMemberForUpdate(ctx, tx, request.GroupID, request.UserID)
		if err != nil && !errors.Is(err, ErrGroupMemberNotFound) {
			return nil, err
		}
		if member != nil && member.Status == model.GroupMemberStatusActive {
			return nil, ErrGroupAlreadyMember
		}

		count, err := countActiveGroupMembers(ctx, tx, request.GroupID)
		if err != nil {
			return nil, err
		}
		if count >= group.MaxMembers {
			return nil, ErrGroupFull
		}
	}

	result, err := tx.ExecContext(ctx, `
UPDATE group_join_requests
SET status = ?, handled_by = ?, updated_at = CURRENT_TIMESTAMP
WHERE request_id = ? AND status = ?
`, nextStatus, operatorID, requestID, model.GroupJoinRequestStatusPending)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrGroupJoinRequestNotFound
	}

	if accept {
		if err := upsertGroupMember(ctx, tx, request.GroupID, request.UserID, model.GroupRoleMember); err != nil {
			return nil, err
		}
		if err := upsertConversationMemberWithRole(ctx, tx, group.ConversationID, request.UserID, model.ConversationMemberRoleMember); err != nil {
			return nil, err
		}
		if err := upsertConversationUserState(ctx, tx, group.ConversationID, request.UserID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &HandleJoinRequestResult{
		GroupID:        request.GroupID,
		ConversationID: group.ConversationID,
		UserID:         request.UserID,
		Status:         nextStatus,
	}, nil
}

func (r *GroupRepository) ListMembers(ctx context.Context, groupID int64, viewerID int64, limit int, offset int) ([]model.GroupMemberWithProfile, int64, error) {
	isMember, err := r.IsActiveMember(ctx, groupID, viewerID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, ErrGroupPermissionDenied
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM group_members
WHERE group_id = ? AND status = ?
`, groupID, model.GroupMemberStatusActive).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.GroupMemberWithProfile{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  gm.id, gm.group_id, gm.user_id, gm.role, gm.mute_until, gm.status, gm.joined_at, gm.left_at,
  CASE
    WHEN gm.user_id = ? THEN 'self'
    WHEN f.id IS NOT NULL THEN 'friend'
    WHEN fr_sent.id IS NOT NULL THEN 'pending_sent'
    WHEN fr_recv.id IS NOT NULL THEN 'pending_received'
    ELSE 'not_friend'
  END AS friendship_status,
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM group_members gm
INNER JOIN users u ON u.user_id = gm.user_id AND u.deleted_at IS NULL
INNER JOIN user_profiles p ON p.user_id = u.user_id
LEFT JOIN friendships f
  ON f.user_id_1 = LEAST(?, gm.user_id)
  AND f.user_id_2 = GREATEST(?, gm.user_id)
  AND f.status = 'normal'
LEFT JOIN friend_requests fr_sent
  ON fr_sent.from_user_id = ?
  AND fr_sent.to_user_id = gm.user_id
  AND fr_sent.status = 'pending'
LEFT JOIN friend_requests fr_recv
  ON fr_recv.from_user_id = gm.user_id
  AND fr_recv.to_user_id = ?
  AND fr_recv.status = 'pending'
WHERE gm.group_id = ?
  AND gm.status = ?
ORDER BY FIELD(gm.role, 'owner', 'admin', 'member'), gm.joined_at ASC
LIMIT ? OFFSET ?
`, viewerID, viewerID, viewerID, viewerID, viewerID, groupID, model.GroupMemberStatusActive, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.GroupMemberWithProfile, 0)
	for rows.Next() {
		item, err := scanGroupMemberWithProfile(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GroupRepository) SetAdmin(ctx context.Context, groupID int64, operatorID int64, targetUserID int64, admin bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}
	operator, err := findGroupMemberForUpdate(ctx, tx, groupID, operatorID)
	if err != nil || operator.Status != model.GroupMemberStatusActive || operator.Role != model.GroupRoleOwner {
		return ErrGroupPermissionDenied
	}
	target, err := findGroupMemberForUpdate(ctx, tx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if target.Status != model.GroupMemberStatusActive || target.Role == model.GroupRoleOwner {
		return ErrGroupMemberNotFound
	}

	nextRole := model.GroupRoleMember
	if admin {
		nextRole = model.GroupRoleAdmin
	}
	if !admin && target.Role != model.GroupRoleAdmin {
		return ErrGroupMemberNotFound
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE group_members
SET role = ?, updated_at = CURRENT_TIMESTAMP
WHERE group_id = ? AND user_id = ? AND status = ?
`, nextRole, groupID, targetUserID, model.GroupMemberStatusActive); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE conversation_members
SET role = ?
WHERE conversation_id = ? AND user_id = ? AND status = ?
`, nextRole, group.ConversationID, targetUserID, model.ConversationMemberStatusActive); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *GroupRepository) SetMute(ctx context.Context, groupID int64, operatorID int64, targetUserID int64, muteUntil sql.NullTime) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}
	operator, err := findGroupMemberForUpdate(ctx, tx, groupID, operatorID)
	if err != nil || operator.Status != model.GroupMemberStatusActive || !isOwnerOrAdmin(operator.Role) {
		return ErrGroupPermissionDenied
	}
	target, err := findGroupMemberForUpdate(ctx, tx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if target.Status != model.GroupMemberStatusActive || target.Role == model.GroupRoleOwner || target.UserID == operatorID {
		return ErrGroupMemberNotFound
	}
	if operator.Role == model.GroupRoleAdmin && target.Role != model.GroupRoleMember {
		return ErrGroupPermissionDenied
	}

	_, err = tx.ExecContext(ctx, `
UPDATE group_members
SET mute_until = ?, updated_at = CURRENT_TIMESTAMP
WHERE group_id = ? AND user_id = ? AND status = ?
`, muteUntil, groupID, targetUserID, model.GroupMemberStatusActive)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *GroupRepository) UpdateSettings(ctx context.Context, groupID int64, operatorID int64, update GroupSettingsUpdate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}
	operator, err := findGroupMemberForUpdate(ctx, tx, groupID, operatorID)
	if err != nil || operator.Status != model.GroupMemberStatusActive {
		return ErrGroupPermissionDenied
	}
	if update.MaxMembers != nil && operator.Role != model.GroupRoleOwner {
		return ErrGroupPermissionDenied
	}
	if update.AllowMemberInvite != nil && !isOwnerOrAdmin(operator.Role) {
		return ErrGroupPermissionDenied
	}

	maxMembers := group.MaxMembers
	if update.MaxMembers != nil {
		maxMembers = *update.MaxMembers
		count, err := countActiveGroupMembers(ctx, tx, groupID)
		if err != nil {
			return err
		}
		if maxMembers < count {
			return ErrGroupFull
		}
	}
	allowMemberInvite := group.AllowMemberInvite
	if update.AllowMemberInvite != nil {
		allowMemberInvite = *update.AllowMemberInvite
	}

	_, err = tx.ExecContext(ctx, `
UPDATE `+"`groups`"+`
SET max_members = ?, allow_member_invite = ?, updated_at = CURRENT_TIMESTAMP
WHERE group_id = ?
`, maxMembers, allowMemberInvite, groupID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *GroupRepository) Dissolve(ctx context.Context, groupID int64, operatorID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}
	operator, err := findGroupMemberForUpdate(ctx, tx, groupID, operatorID)
	if err != nil || operator.Status != model.GroupMemberStatusActive || operator.Role != model.GroupRoleOwner {
		return ErrGroupPermissionDenied
	}

	_, err = tx.ExecContext(ctx, `
UPDATE `+"`groups`"+`
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE group_id = ? AND status = ?
`, model.GroupStatusDissolved, groupID, model.GroupStatusNormal)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *GroupRepository) Leave(ctx context.Context, groupID int64, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	group, err := findGroupByIDForUpdate(ctx, tx, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}

	member, err := findGroupMemberForUpdate(ctx, tx, groupID, userID)
	if err != nil {
		return err
	}
	if member.Status != model.GroupMemberStatusActive {
		return ErrGroupMemberNotFound
	}
	if member.Role == model.GroupRoleOwner {
		return ErrGroupOwnerCannotLeave
	}

	leftAt := time.Now()
	if _, err := tx.ExecContext(ctx, `
UPDATE group_members
SET status = ?,
    mute_until = NULL,
    left_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE group_id = ?
  AND user_id = ?
  AND status = ?
`, model.GroupMemberStatusLeft, leftAt, groupID, userID, model.GroupMemberStatusActive); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE conversation_members
SET status = ?,
    left_at = ?
WHERE conversation_id = ?
  AND user_id = ?
  AND status = ?
`, model.ConversationMemberStatusLeft, leftAt, group.ConversationID, userID, model.ConversationMemberStatusActive); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE conversation_user_states
SET is_deleted = 1,
    cleared_at = ?,
    unread_count = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ?
  AND user_id = ?
`, leftAt, group.ConversationID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *GroupRepository) CountActiveMembers(ctx context.Context, groupID int64) (int, error) {
	return countActiveGroupMembers(ctx, r.db, groupID)
}

func (r *GroupRepository) IsActiveMember(ctx context.Context, groupID int64, userID int64) (bool, error) {
	member, err := r.FindMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, ErrGroupMemberNotFound) {
			return false, nil
		}
		return false, err
	}
	return member.Status == model.GroupMemberStatusActive, nil
}

func (r *GroupRepository) FindMember(ctx context.Context, groupID int64, userID int64) (*model.GroupMember, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, group_id, user_id, role, mute_until, status, joined_at, left_at
FROM group_members
WHERE group_id = ? AND user_id = ?
LIMIT 1
`, groupID, userID)
	member, err := scanGroupMember(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupMemberNotFound
		}
		return nil, err
	}
	return member, nil
}

func (r *GroupRepository) FindMessageContextByConversation(ctx context.Context, conversationID int64, userID int64) (*GroupMessageContext, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT
  g.id, g.group_id, g.group_no, g.conversation_id, g.owner_id, g.name, g.avatar_url,
  g.max_members, g.allow_member_invite, g.status, g.created_at, g.updated_at,
  gm.id, gm.group_id, gm.user_id, gm.role, gm.mute_until, gm.status, gm.joined_at, gm.left_at,
  p.nickname, p.avatar_url
FROM `+"`groups`"+` g
INNER JOIN group_members gm ON gm.group_id = g.group_id AND gm.user_id = ?
INNER JOIN user_profiles p ON p.user_id = gm.user_id
WHERE g.conversation_id = ?
LIMIT 1
`, userID, conversationID)

	var ctxValue GroupMessageContext
	var avatar sql.NullString
	err := row.Scan(
		&ctxValue.Group.ID,
		&ctxValue.Group.GroupID,
		&ctxValue.Group.GroupNo,
		&ctxValue.Group.ConversationID,
		&ctxValue.Group.OwnerID,
		&ctxValue.Group.Name,
		&ctxValue.Group.AvatarURL,
		&ctxValue.Group.MaxMembers,
		&ctxValue.Group.AllowMemberInvite,
		&ctxValue.Group.Status,
		&ctxValue.Group.CreatedAt,
		&ctxValue.Group.UpdatedAt,
		&ctxValue.Member.ID,
		&ctxValue.Member.GroupID,
		&ctxValue.Member.UserID,
		&ctxValue.Member.Role,
		&ctxValue.Member.MuteUntil,
		&ctxValue.Member.Status,
		&ctxValue.Member.JoinedAt,
		&ctxValue.Member.LeftAt,
		&ctxValue.Sender.Nickname,
		&avatar,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupMemberNotFound
		}
		return nil, err
	}
	ctxValue.Sender.UserID = userID
	ctxValue.Sender.AvatarURL = avatar.String
	ctxValue.Sender.Role = ctxValue.Member.Role
	return &ctxValue, nil
}

func (r *GroupRepository) ListMessageSenders(ctx context.Context, groupID int64, userIDs []int64) (map[int64]model.GroupMessageSender, error) {
	result := make(map[int64]model.GroupMessageSender)
	if len(userIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(userIDs))
	args := make([]any, 0, len(userIDs)+1)
	args = append(args, groupID)
	for _, userID := range userIDs {
		placeholders = append(placeholders, "?")
		args = append(args, userID)
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT gm.user_id, p.nickname, p.avatar_url, gm.role
FROM group_members gm
INNER JOIN user_profiles p ON p.user_id = gm.user_id
WHERE gm.group_id = ?
  AND gm.user_id IN (`+strings.Join(placeholders, ",")+`)
`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sender model.GroupMessageSender
		var avatar sql.NullString
		if err := rows.Scan(&sender.UserID, &sender.Nickname, &avatar, &sender.Role); err != nil {
			return nil, err
		}
		sender.AvatarURL = avatar.String
		if sender.Role == "" {
			sender.Role = model.GroupRoleMember
		}
		result[sender.UserID] = sender
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *GroupRepository) requireRole(ctx context.Context, groupID int64, userID int64, allowed ...string) error {
	group, err := findGroupByID(ctx, r.db, groupID)
	if err != nil {
		return err
	}
	if group.Status != model.GroupStatusNormal {
		return ErrGroupDissolved
	}
	member, err := r.FindMember(ctx, groupID, userID)
	if err != nil {
		return ErrGroupPermissionDenied
	}
	if member.Status != model.GroupMemberStatusActive {
		return ErrGroupPermissionDenied
	}
	for _, role := range allowed {
		if member.Role == role {
			return nil
		}
	}
	return ErrGroupPermissionDenied
}

func findGroupByID(ctx context.Context, exec Executor, groupID int64) (*model.Group, error) {
	row := exec.QueryRowContext(ctx, `
SELECT id, group_id, group_no, conversation_id, owner_id, name, avatar_url,
  max_members, allow_member_invite, status, created_at, updated_at
FROM `+"`groups`"+`
WHERE group_id = ?
LIMIT 1
`, groupID)
	group, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return group, nil
}

func findGroupByIDForUpdate(ctx context.Context, exec Executor, groupID int64) (*model.Group, error) {
	row := exec.QueryRowContext(ctx, `
SELECT id, group_id, group_no, conversation_id, owner_id, name, avatar_url,
  max_members, allow_member_invite, status, created_at, updated_at
FROM `+"`groups`"+`
WHERE group_id = ?
LIMIT 1
FOR UPDATE
`, groupID)
	group, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return group, nil
}

func findGroupMemberForUpdate(ctx context.Context, exec Executor, groupID int64, userID int64) (*model.GroupMember, error) {
	row := exec.QueryRowContext(ctx, `
SELECT id, group_id, user_id, role, mute_until, status, joined_at, left_at
FROM group_members
WHERE group_id = ? AND user_id = ?
LIMIT 1
FOR UPDATE
`, groupID, userID)
	member, err := scanGroupMember(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupMemberNotFound
		}
		return nil, err
	}
	return member, nil
}

func findJoinRequestForUpdate(ctx context.Context, exec Executor, requestID int64) (*model.GroupJoinRequest, error) {
	row := exec.QueryRowContext(ctx, `
SELECT id, request_id, group_id, user_id, message, status, handled_by, created_at, updated_at
FROM group_join_requests
WHERE request_id = ?
LIMIT 1
FOR UPDATE
`, requestID)
	request, err := scanGroupJoinRequest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupJoinRequestNotFound
		}
		return nil, err
	}
	return request, nil
}

func countActiveGroupMembers(ctx context.Context, exec Executor, groupID int64) (int, error) {
	var count int
	err := exec.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM group_members
WHERE group_id = ? AND status = ?
`, groupID, model.GroupMemberStatusActive).Scan(&count)
	return count, err
}

func upsertGroupMember(ctx context.Context, exec Executor, groupID int64, userID int64, role string) error {
	now := time.Now()
	_, err := exec.ExecContext(ctx, `
INSERT INTO group_members (group_id, user_id, role, status, mute_until, joined_at, left_at)
VALUES (?, ?, ?, ?, NULL, ?, NULL)
ON DUPLICATE KEY UPDATE role = VALUES(role), status = VALUES(status), mute_until = NULL, joined_at = ?, left_at = NULL, updated_at = CURRENT_TIMESTAMP
`, groupID, userID, role, model.GroupMemberStatusActive, now, now)
	return err
}

func upsertConversationMemberWithRole(ctx context.Context, exec Executor, conversationID int64, userID int64, role string) error {
	now := time.Now()
	_, err := exec.ExecContext(ctx, `
INSERT INTO conversation_members (conversation_id, user_id, role, status, joined_at)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE role = VALUES(role), status = VALUES(status), joined_at = ?, left_at = NULL
`, conversationID, userID, role, model.ConversationMemberStatusActive, now, now)
	return err
}

func isOwnerOrAdmin(role string) bool {
	return role == model.GroupRoleOwner || role == model.GroupRoleAdmin
}

func scanGroup(row scanner) (*model.Group, error) {
	var group model.Group
	if err := row.Scan(
		&group.ID,
		&group.GroupID,
		&group.GroupNo,
		&group.ConversationID,
		&group.OwnerID,
		&group.Name,
		&group.AvatarURL,
		&group.MaxMembers,
		&group.AllowMemberInvite,
		&group.Status,
		&group.CreatedAt,
		&group.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &group, nil
}

func scanGroupMember(row scanner) (*model.GroupMember, error) {
	var member model.GroupMember
	if err := row.Scan(
		&member.ID,
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.MuteUntil,
		&member.Status,
		&member.JoinedAt,
		&member.LeftAt,
	); err != nil {
		return nil, err
	}
	return &member, nil
}

func scanGroupJoinRequest(row scanner) (*model.GroupJoinRequest, error) {
	var request model.GroupJoinRequest
	if err := row.Scan(
		&request.ID,
		&request.RequestID,
		&request.GroupID,
		&request.UserID,
		&request.Message,
		&request.Status,
		&request.HandledBy,
		&request.CreatedAt,
		&request.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &request, nil
}

func scanGroupJoinRequestWithUser(row scanner) (*model.GroupJoinRequestWithUser, error) {
	var item model.GroupJoinRequestWithUser
	if err := row.Scan(
		&item.Request.ID,
		&item.Request.RequestID,
		&item.Request.GroupID,
		&item.Request.UserID,
		&item.Request.Message,
		&item.Request.Status,
		&item.Request.HandledBy,
		&item.Request.CreatedAt,
		&item.Request.UpdatedAt,
		&item.User.User.ID,
		&item.User.User.UserID,
		&item.User.User.Username,
		&item.User.User.PasswordHash,
		&item.User.User.UserType,
		&item.User.User.Status,
		&item.User.User.CreatedAt,
		&item.User.User.UpdatedAt,
		&item.User.User.DeletedAt,
		&item.User.Profile.ID,
		&item.User.Profile.UserID,
		&item.User.Profile.Nickname,
		&item.User.Profile.AvatarURL,
		&item.User.Profile.Gender,
		&item.User.Profile.Bio,
		&item.User.Profile.ProfileStatus,
		&item.User.Profile.ProfileReviewReason,
		&item.User.Profile.CreatedAt,
		&item.User.Profile.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanGroupMemberWithProfile(row scanner) (*model.GroupMemberWithProfile, error) {
	var item model.GroupMemberWithProfile
	if err := row.Scan(
		&item.Member.ID,
		&item.Member.GroupID,
		&item.Member.UserID,
		&item.Member.Role,
		&item.Member.MuteUntil,
		&item.Member.Status,
		&item.Member.JoinedAt,
		&item.Member.LeftAt,
		&item.FriendshipStatus,
		&item.User.User.ID,
		&item.User.User.UserID,
		&item.User.User.Username,
		&item.User.User.PasswordHash,
		&item.User.User.UserType,
		&item.User.User.Status,
		&item.User.User.CreatedAt,
		&item.User.User.UpdatedAt,
		&item.User.User.DeletedAt,
		&item.User.Profile.ID,
		&item.User.Profile.UserID,
		&item.User.Profile.Nickname,
		&item.User.Profile.AvatarURL,
		&item.User.Profile.Gender,
		&item.User.Profile.Bio,
		&item.User.Profile.ProfileStatus,
		&item.User.Profile.ProfileReviewReason,
		&item.User.Profile.CreatedAt,
		&item.User.Profile.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}
