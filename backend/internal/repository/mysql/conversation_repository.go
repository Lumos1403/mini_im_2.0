package mysql

import (
	"context"
	"database/sql"
	"errors"

	"mini_im/backend/internal/model"
)

var ErrConversationNotFound = errors.New("conversation not found")

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) EnsurePrivateConversation(ctx context.Context, conversationID int64, userID1 int64, userID2 int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	existingConversationID, err := r.EnsurePrivateConversationInTx(ctx, tx, conversationID, userID1, userID2)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return existingConversationID, nil
}

func (r *ConversationRepository) EnsurePrivateConversationInTx(ctx context.Context, exec Executor, conversationID int64, userID1 int64, userID2 int64) (int64, error) {
	if userID1 <= 0 || userID2 <= 0 || userID1 == userID2 {
		return 0, ErrConversationNotFound
	}

	existingConversationID, err := findPrivateConversationID(ctx, exec, userID1, userID2)
	if err != nil && !errors.Is(err, ErrConversationNotFound) {
		return 0, err
	}
	if errors.Is(err, ErrConversationNotFound) {
		existingConversationID = conversationID
		if _, err := exec.ExecContext(ctx, `
INSERT INTO conversations (conversation_id, conversation_type, status)
VALUES (?, ?, ?)
`, existingConversationID, model.ConversationTypePrivate, model.ConversationStatusNormal); err != nil {
			return 0, err
		}
	} else if _, err := exec.ExecContext(ctx, `
UPDATE conversations
SET status = ?
WHERE conversation_id = ?
`, model.ConversationStatusNormal, existingConversationID); err != nil {
		return 0, err
	}

	if err := upsertConversationMember(ctx, exec, existingConversationID, userID1); err != nil {
		return 0, err
	}
	if err := upsertConversationMember(ctx, exec, existingConversationID, userID2); err != nil {
		return 0, err
	}
	if err := upsertConversationUserState(ctx, exec, existingConversationID, userID1); err != nil {
		return 0, err
	}
	if err := upsertConversationUserState(ctx, exec, existingConversationID, userID2); err != nil {
		return 0, err
	}

	return existingConversationID, nil
}

func (r *ConversationRepository) ListUserConversations(ctx context.Context, userID int64, limit int, offset int) ([]model.ConversationListItem, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM conversations c
INNER JOIN conversation_user_states s ON s.conversation_id = c.conversation_id AND s.user_id = ?
INNER JOIN conversation_members self_cm ON self_cm.conversation_id = c.conversation_id AND self_cm.user_id = ?
WHERE c.status = ?
  AND s.is_deleted = 0
  AND self_cm.status = ?
`, userID, userID, model.ConversationStatusNormal, model.ConversationMemberStatusActive).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.ConversationListItem{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  c.conversation_id, c.conversation_type, c.last_message_id, c.last_message_at, c.updated_at,
  s.unread_count, s.is_pinned, s.is_muted,
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM conversations c
INNER JOIN conversation_user_states s ON s.conversation_id = c.conversation_id AND s.user_id = ?
INNER JOIN conversation_members self_cm ON self_cm.conversation_id = c.conversation_id AND self_cm.user_id = ?
LEFT JOIN conversation_members peer_cm
  ON peer_cm.conversation_id = c.conversation_id
  AND peer_cm.user_id <> ?
  AND c.conversation_type = ?
LEFT JOIN users u ON u.user_id = peer_cm.user_id AND u.deleted_at IS NULL
LEFT JOIN user_profiles p ON p.user_id = u.user_id
WHERE c.status = ?
  AND s.is_deleted = 0
  AND self_cm.status = ?
ORDER BY s.is_pinned DESC, COALESCE(c.last_message_at, c.updated_at) DESC, c.updated_at DESC
LIMIT ? OFFSET ?
`, userID, userID, userID, model.ConversationTypePrivate, model.ConversationStatusNormal, model.ConversationMemberStatusActive, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.ConversationListItem, 0)
	for rows.Next() {
		item, err := scanConversationListItem(rows)
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

func (r *ConversationRepository) FindPrivatePeerID(ctx context.Context, conversationID int64, userID int64) (int64, error) {
	var peerID int64
	err := r.db.QueryRowContext(ctx, `
SELECT peer_cm.user_id
FROM conversations c
INNER JOIN conversation_members self_cm
  ON self_cm.conversation_id = c.conversation_id
  AND self_cm.user_id = ?
  AND self_cm.status = ?
INNER JOIN conversation_members peer_cm
  ON peer_cm.conversation_id = c.conversation_id
  AND peer_cm.user_id <> ?
  AND peer_cm.status = ?
WHERE c.conversation_id = ?
  AND c.conversation_type = ?
  AND c.status = ?
LIMIT 1
`, userID, model.ConversationMemberStatusActive, userID, model.ConversationMemberStatusActive, conversationID, model.ConversationTypePrivate, model.ConversationStatusNormal).Scan(&peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrConversationNotFound
		}
		return 0, err
	}
	return peerID, nil
}

func findPrivateConversationID(ctx context.Context, exec Executor, userID1 int64, userID2 int64) (int64, error) {
	var conversationID int64
	err := exec.QueryRowContext(ctx, `
SELECT c.conversation_id
FROM conversations c
INNER JOIN conversation_members cm1 ON cm1.conversation_id = c.conversation_id AND cm1.user_id = ?
INNER JOIN conversation_members cm2 ON cm2.conversation_id = c.conversation_id AND cm2.user_id = ?
WHERE c.conversation_type = ?
LIMIT 1
FOR UPDATE
`, userID1, userID2, model.ConversationTypePrivate).Scan(&conversationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrConversationNotFound
		}
		return 0, err
	}
	return conversationID, nil
}

func upsertConversationMember(ctx context.Context, exec Executor, conversationID int64, userID int64) error {
	_, err := exec.ExecContext(ctx, `
INSERT INTO conversation_members (conversation_id, user_id, role, status)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE role = VALUES(role), status = VALUES(status), left_at = NULL
`, conversationID, userID, model.ConversationMemberRoleMember, model.ConversationMemberStatusActive)
	return err
}

func upsertConversationUserState(ctx context.Context, exec Executor, conversationID int64, userID int64) error {
	_, err := exec.ExecContext(ctx, `
INSERT INTO conversation_user_states (conversation_id, user_id, is_deleted, unread_count, is_pinned, is_muted)
VALUES (?, ?, 0, 0, 0, 0)
ON DUPLICATE KEY UPDATE is_deleted = 0, updated_at = CURRENT_TIMESTAMP
`, conversationID, userID)
	return err
}

func scanConversationListItem(row scanner) (*model.ConversationListItem, error) {
	var item model.ConversationListItem
	var peer peerUserWithProfileScan
	if err := row.Scan(
		&item.ConversationID,
		&item.ConversationType,
		&item.LastMessageID,
		&item.LastMessageAt,
		&item.UpdatedAt,
		&item.UnreadCount,
		&item.IsPinned,
		&item.IsMuted,
		&peer.UserID,
		&peer.PublicUserID,
		&peer.Username,
		&peer.PasswordHash,
		&peer.UserType,
		&peer.UserStatus,
		&peer.UserCreatedAt,
		&peer.UserUpdatedAt,
		&peer.DeletedAt,
		&peer.ProfileID,
		&peer.ProfileUserID,
		&peer.Nickname,
		&peer.AvatarURL,
		&peer.Gender,
		&peer.Bio,
		&peer.ProfileStatus,
		&peer.ProfileReviewReason,
		&peer.ProfileCreatedAt,
		&peer.ProfileUpdatedAt,
	); err != nil {
		return nil, err
	}

	if peer.PublicUserID.Valid {
		item.Peer = peer.toModel()
	}
	return &item, nil
}

type peerUserWithProfileScan struct {
	UserID              sql.NullInt64
	PublicUserID        sql.NullInt64
	Username            sql.NullString
	PasswordHash        sql.NullString
	UserType            sql.NullString
	UserStatus          sql.NullString
	UserCreatedAt       sql.NullTime
	UserUpdatedAt       sql.NullTime
	DeletedAt           sql.NullTime
	ProfileID           sql.NullInt64
	ProfileUserID       sql.NullInt64
	Nickname            sql.NullString
	AvatarURL           sql.NullString
	Gender              sql.NullString
	Bio                 sql.NullString
	ProfileStatus       sql.NullString
	ProfileReviewReason sql.NullString
	ProfileCreatedAt    sql.NullTime
	ProfileUpdatedAt    sql.NullTime
}

func (p peerUserWithProfileScan) toModel() model.UserWithProfile {
	return model.UserWithProfile{
		User: model.User{
			ID:           p.UserID.Int64,
			UserID:       p.PublicUserID.Int64,
			Username:     p.Username.String,
			PasswordHash: p.PasswordHash.String,
			UserType:     p.UserType.String,
			Status:       p.UserStatus.String,
			CreatedAt:    p.UserCreatedAt.Time,
			UpdatedAt:    p.UserUpdatedAt.Time,
			DeletedAt:    p.DeletedAt,
		},
		Profile: model.UserProfile{
			ID:                  p.ProfileID.Int64,
			UserID:              p.ProfileUserID.Int64,
			Nickname:            p.Nickname.String,
			AvatarURL:           p.AvatarURL,
			Gender:              p.Gender,
			Bio:                 p.Bio,
			ProfileStatus:       p.ProfileStatus.String,
			ProfileReviewReason: p.ProfileReviewReason,
			CreatedAt:           p.ProfileCreatedAt.Time,
			UpdatedAt:           p.ProfileUpdatedAt.Time,
		},
	}
}
