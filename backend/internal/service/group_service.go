package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
)

const maxGroupNoRetries = 12

type GroupService struct {
	groupRepo       *mysqlrepo.GroupRepository
	idGenerator     *snowflake.Node
	groupMaxMembers int
}

func NewGroupService(groupRepo *mysqlrepo.GroupRepository, idGenerator *snowflake.Node, groupMaxMembers int) *GroupService {
	if groupMaxMembers <= 0 {
		groupMaxMembers = 50
	}
	return &GroupService{
		groupRepo:       groupRepo,
		idGenerator:     idGenerator,
		groupMaxMembers: groupMaxMembers,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, ownerID int64, input CreateGroupInput) (*CreateGroupOutput, *apperrors.AppError) {
	name := strings.TrimSpace(input.Name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return nil, apperrors.ErrInvalidParam
	}
	avatarURL := strings.TrimSpace(input.AvatarURL)
	if utf8.RuneCountInString(avatarURL) > 512 {
		return nil, apperrors.ErrInvalidParam
	}

	var lastErr error
	for attempt := 0; attempt < maxGroupNoRetries; attempt++ {
		group := &model.Group{
			GroupID:           s.idGenerator.NextID(),
			GroupNo:           randomGroupNo(),
			ConversationID:    s.idGenerator.NextID(),
			OwnerID:           ownerID,
			Name:              name,
			AvatarURL:         sql.NullString{String: avatarURL, Valid: avatarURL != ""},
			MaxMembers:        s.groupMaxMembers,
			AllowMemberInvite: true,
			Status:            model.GroupStatusNormal,
			CreatedAt:         now(),
		}

		err := s.groupRepo.CreateGroupWithOwner(ctx, group)
		if err == nil {
			return &CreateGroupOutput{
				GroupID:        formatID(group.GroupID),
				GroupNo:        group.GroupNo,
				ConversationID: formatID(group.ConversationID),
			}, nil
		}
		lastErr = err
		if !errors.Is(err, mysqlrepo.ErrDuplicateGroupNo) {
			break
		}
	}
	if errors.Is(lastErr, mysqlrepo.ErrDuplicateGroupNo) {
		return nil, apperrors.ErrInternal
	}
	return nil, mapGroupRepositoryError(lastErr)
}

func (s *GroupService) SearchGroups(ctx context.Context, userID int64, keyword string) (*PageOutput[GroupOutput], *apperrors.AppError) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || utf8.RuneCountInString(keyword) > 100 {
		return nil, apperrors.ErrInvalidParam
	}

	groups, err := s.groupRepo.Search(ctx, keyword, userID, 20)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	outputs := make([]GroupOutput, 0, len(groups))
	for i := range groups {
		isMember, err := s.groupRepo.IsActiveMember(ctx, groups[i].GroupID, userID)
		if err != nil {
			return nil, apperrors.ErrInternal
		}
		output := toGroupOutput(&groups[i])
		output.IsMember = isMember
		outputs = append(outputs, output)
	}
	return &PageOutput[GroupOutput]{
		List:     outputs,
		Total:    int64(len(outputs)),
		Page:     1,
		PageSize: 20,
	}, nil
}

func (s *GroupService) CreateJoinRequest(ctx context.Context, userID int64, groupIDValue string, input CreateGroupJoinRequestInput) (*CreateGroupJoinRequestOutput, *apperrors.AppError) {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return nil, appErr
	}
	message := strings.TrimSpace(input.Message)
	if utf8.RuneCountInString(message) > 255 {
		return nil, apperrors.ErrInvalidParam
	}

	requestID := s.idGenerator.NextID()
	request := &model.GroupJoinRequest{
		RequestID: requestID,
		GroupID:   groupID,
		UserID:    userID,
		Message:   sql.NullString{String: message, Valid: message != ""},
		Status:    model.GroupJoinRequestStatusPending,
	}
	if err := s.groupRepo.CreateJoinRequest(ctx, request); err != nil {
		return nil, mapGroupRepositoryError(err)
	}

	return &CreateGroupJoinRequestOutput{
		RequestID: formatID(requestID),
		Status:    model.GroupJoinRequestStatusPending,
	}, nil
}

func (s *GroupService) ListJoinRequests(ctx context.Context, userID int64, groupIDValue string, page int, pageSize int) (*PageOutput[GroupJoinRequestOutput], *apperrors.AppError) {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return nil, appErr
	}
	page, pageSize, offset := normalizePagination(page, pageSize)
	items, total, err := s.groupRepo.ListJoinRequests(ctx, groupID, userID, pageSize, offset)
	if err != nil {
		return nil, mapGroupRepositoryError(err)
	}

	outputs := make([]GroupJoinRequestOutput, 0, len(items))
	for i := range items {
		outputs = append(outputs, toGroupJoinRequestOutput(&items[i]))
	}
	return &PageOutput[GroupJoinRequestOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GroupService) AcceptJoinRequest(ctx context.Context, userID int64, requestIDValue string) (*HandleGroupJoinRequestOutput, *apperrors.AppError) {
	return s.handleJoinRequest(ctx, userID, requestIDValue, true)
}

func (s *GroupService) RejectJoinRequest(ctx context.Context, userID int64, requestIDValue string) (*HandleGroupJoinRequestOutput, *apperrors.AppError) {
	return s.handleJoinRequest(ctx, userID, requestIDValue, false)
}

func (s *GroupService) handleJoinRequest(ctx context.Context, userID int64, requestIDValue string, accept bool) (*HandleGroupJoinRequestOutput, *apperrors.AppError) {
	requestID, appErr := parsePositiveID(requestIDValue)
	if appErr != nil {
		return nil, appErr
	}
	result, err := s.groupRepo.HandleJoinRequest(ctx, requestID, userID, accept)
	if err != nil {
		return nil, mapGroupRepositoryError(err)
	}
	return &HandleGroupJoinRequestOutput{
		RequestID:      formatID(requestID),
		GroupID:        formatID(result.GroupID),
		ConversationID: formatID(result.ConversationID),
		UserID:         formatID(result.UserID),
		Status:         result.Status,
	}, nil
}

func (s *GroupService) ListMembers(ctx context.Context, userID int64, groupIDValue string, page int, pageSize int) (*PageOutput[GroupMemberOutput], *apperrors.AppError) {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return nil, appErr
	}
	page, pageSize, offset := normalizePagination(page, pageSize)
	items, total, err := s.groupRepo.ListMembers(ctx, groupID, userID, pageSize, offset)
	if err != nil {
		return nil, mapGroupRepositoryError(err)
	}

	outputs := make([]GroupMemberOutput, 0, len(items))
	for i := range items {
		outputs = append(outputs, toGroupMemberOutput(&items[i]))
	}
	return &PageOutput[GroupMemberOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GroupService) SetAdmin(ctx context.Context, userID int64, groupIDValue string, targetUserIDValue string, admin bool) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	targetUserID, appErr := parsePositiveID(targetUserIDValue)
	if appErr != nil {
		return appErr
	}
	if targetUserID == userID {
		return apperrors.ErrInvalidParam
	}
	if err := s.groupRepo.SetAdmin(ctx, groupID, userID, targetUserID, admin); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func (s *GroupService) MuteMember(ctx context.Context, userID int64, groupIDValue string, targetUserIDValue string, input MuteGroupMemberInput) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	targetUserID, appErr := parsePositiveID(targetUserIDValue)
	if appErr != nil {
		return appErr
	}
	muteUntil, appErr := parseMuteUntil(input.MuteUntil)
	if appErr != nil {
		return appErr
	}
	if err := s.groupRepo.SetMute(ctx, groupID, userID, targetUserID, sql.NullTime{Time: muteUntil, Valid: true}); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func (s *GroupService) UnmuteMember(ctx context.Context, userID int64, groupIDValue string, targetUserIDValue string) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	targetUserID, appErr := parsePositiveID(targetUserIDValue)
	if appErr != nil {
		return appErr
	}
	if err := s.groupRepo.SetMute(ctx, groupID, userID, targetUserID, sql.NullTime{}); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func (s *GroupService) UpdateSettings(ctx context.Context, userID int64, groupIDValue string, input UpdateGroupSettingsInput) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	if input.AllowMemberInvite == nil && input.MaxMembers == nil {
		return apperrors.ErrInvalidParam
	}
	if input.MaxMembers != nil {
		if *input.MaxMembers <= 0 || *input.MaxMembers > s.groupMaxMembers {
			return apperrors.ErrInvalidParam
		}
	}

	update := mysqlrepo.GroupSettingsUpdate{
		AllowMemberInvite: input.AllowMemberInvite,
		MaxMembers:        input.MaxMembers,
	}
	if err := s.groupRepo.UpdateSettings(ctx, groupID, userID, update); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func (s *GroupService) DissolveGroup(ctx context.Context, userID int64, groupIDValue string) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	if err := s.groupRepo.Dissolve(ctx, groupID, userID); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func (s *GroupService) LeaveGroup(ctx context.Context, userID int64, groupIDValue string) *apperrors.AppError {
	groupID, appErr := parsePositiveID(groupIDValue)
	if appErr != nil {
		return appErr
	}
	if err := s.groupRepo.Leave(ctx, groupID, userID); err != nil {
		return mapGroupRepositoryError(err)
	}
	return nil
}

func randomGroupNo() string {
	length := 8 + randomInt(3)
	digits := make([]byte, length)
	digits[0] = byte('1' + randomInt(9))
	for i := 1; i < length; i++ {
		digits[i] = byte('0' + randomInt(10))
	}
	return string(digits)
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return int(time.Now().UnixNano() % int64(max))
	}
	return int(value.Int64())
}

func parseMuteUntil(value string) (time.Time, *apperrors.AppError) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, apperrors.ErrInvalidParam
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			if !parsed.After(now()) {
				return time.Time{}, apperrors.ErrInvalidParam
			}
			return parsed, nil
		}
	}
	return time.Time{}, apperrors.ErrInvalidParam
}

func toGroupOutput(group *model.Group) GroupOutput {
	return GroupOutput{
		GroupID:           formatID(group.GroupID),
		GroupNo:           group.GroupNo,
		ConversationID:    formatID(group.ConversationID),
		OwnerID:           formatID(group.OwnerID),
		Name:              group.Name,
		AvatarURL:         group.AvatarURL.String,
		MaxMembers:        group.MaxMembers,
		AllowMemberInvite: group.AllowMemberInvite,
		Status:            group.Status,
	}
}

func toGroupJoinRequestOutput(item *model.GroupJoinRequestWithUser) GroupJoinRequestOutput {
	return GroupJoinRequestOutput{
		RequestID: formatID(item.Request.RequestID),
		GroupID:   formatID(item.Request.GroupID),
		UserID:    formatID(item.Request.UserID),
		User:      toUserOutput(&item.User),
		Message:   item.Request.Message.String,
		Status:    item.Request.Status,
		HandledBy: formatOptionalID(item.Request.HandledBy),
		CreatedAt: item.Request.CreatedAt,
		UpdatedAt: item.Request.UpdatedAt,
	}
}

func toGroupMemberOutput(item *model.GroupMemberWithProfile) GroupMemberOutput {
	var muteUntil *string
	if item.Member.MuteUntil.Valid {
		value := formatTime(item.Member.MuteUntil.Time)
		muteUntil = &value
	}
	return GroupMemberOutput{
		UserID:           formatID(item.Member.UserID),
		Nickname:         item.User.Profile.Nickname,
		AvatarURL:        item.User.Profile.AvatarURL.String,
		Bio:              item.User.Profile.Bio.String,
		Role:             item.Member.Role,
		MuteUntil:        muteUntil,
		JoinedAt:         formatTime(item.Member.JoinedAt),
		Status:           item.Member.Status,
		FriendshipStatus: item.FriendshipStatus,
	}
}

func mapGroupRepositoryError(err error) *apperrors.AppError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mysqlrepo.ErrGroupNotFound):
		return apperrors.ErrGroupNotFound
	case errors.Is(err, mysqlrepo.ErrGroupPermissionDenied):
		return apperrors.ErrGroupPermissionDenied
	case errors.Is(err, mysqlrepo.ErrGroupJoinRequestPending):
		return apperrors.ErrGroupJoinPending
	case errors.Is(err, mysqlrepo.ErrGroupJoinRequestNotFound):
		return apperrors.ErrGroupJoinNotFound
	case errors.Is(err, mysqlrepo.ErrGroupAlreadyMember):
		return apperrors.ErrGroupAlreadyMember
	case errors.Is(err, mysqlrepo.ErrGroupFull):
		return apperrors.ErrGroupFull
	case errors.Is(err, mysqlrepo.ErrGroupDissolved):
		return apperrors.ErrGroupDissolved
	case errors.Is(err, mysqlrepo.ErrGroupMemberNotFound):
		return apperrors.ErrGroupMemberNotFound
	case errors.Is(err, mysqlrepo.ErrGroupMemberMuted):
		return apperrors.ErrGroupMemberMuted
	case errors.Is(err, mysqlrepo.ErrGroupOwnerCannotLeave):
		return apperrors.ErrGroupOwnerCannotLeave
	default:
		return apperrors.ErrInternal
	}
}
