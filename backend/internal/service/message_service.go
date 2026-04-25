package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
	redisrepo "mini_im/backend/internal/repository/redis"
)

const (
	wsFailureInvalidRequest               = "invalid_request"
	wsFailureConversationNotFound         = "conversation_not_found"
	wsFailureNotFriends                   = "not_friends"
	wsFailureBlocked                      = "failed_blocked"
	wsFailureDuplicateClientMsgIDConflict = "duplicate_client_msg_id_conflict"
	wsFailureFileNotFound                 = "file_not_found"
	wsFailureFileAccessDenied             = "file_access_denied"
	wsFailureInternal                     = "internal_error"
)

type SendMessageInput struct {
	SenderID       int64
	ConversationID string
	ClientMsgID    string
	MessageType    string
	Content        string
	ExtraJSON      json.RawMessage
}

type normalizedMessagePayload struct {
	ConversationID int64
	ClientMsgID    string
	MessageType    string
	Content        string
	ExtraJSON      string
}

type SendMessageResult struct {
	Ack        *SendMessageAckOutput
	Failed     *SendMessageFailedOutput
	Receive    *MessageReceiveOutput
	ReceiverID int64
	Duplicated bool
}

type MessageService struct {
	conversationRepo     *mysqlrepo.ConversationRepository
	friendRepo           *mysqlrepo.FriendRepository
	messageRepo          *mysqlrepo.MessageRepository
	fileRepo             *mysqlrepo.FileRepository
	messageCacheRepo     *redisrepo.MessageRepository
	idGenerator          *snowflake.Node
	textMessageMaxLength int
	recallWindow         time.Duration
	recallNotifier       MessageRecallNotifier
}

type MessageRecallNotifier interface {
	NotifyMessageRecalled(ctx context.Context, recipientIDs []int64, data MessageRecalledEventOutput) error
}

func NewMessageService(
	conversationRepo *mysqlrepo.ConversationRepository,
	friendRepo *mysqlrepo.FriendRepository,
	messageRepo *mysqlrepo.MessageRepository,
	fileRepo *mysqlrepo.FileRepository,
	messageCacheRepo *redisrepo.MessageRepository,
	idGenerator *snowflake.Node,
	textMessageMaxLength int,
	recallMinutes int,
) *MessageService {
	if textMessageMaxLength <= 0 {
		textMessageMaxLength = 2000
	}
	if recallMinutes <= 0 {
		recallMinutes = 5
	}
	return &MessageService{
		conversationRepo:     conversationRepo,
		friendRepo:           friendRepo,
		messageRepo:          messageRepo,
		fileRepo:             fileRepo,
		messageCacheRepo:     messageCacheRepo,
		idGenerator:          idGenerator,
		textMessageMaxLength: textMessageMaxLength,
		recallWindow:         time.Duration(recallMinutes) * time.Minute,
	}
}

func (s *MessageService) SetRecallNotifier(notifier MessageRecallNotifier) {
	s.recallNotifier = notifier
}

func (s *MessageService) SendMessage(ctx context.Context, input SendMessageInput) (*SendMessageResult, *apperrors.AppError) {
	payload, appErr := s.validateSendInput(ctx, input)
	if appErr != nil {
		if appErr == apperrors.ErrInternal {
			return nil, appErr
		}
		return &SendMessageResult{Failed: failedMessage(input.ClientMsgID, input.ConversationID, model.MessageSendStatusFailed, sendValidationFailureCode(appErr), appErr.Message)}, nil
	}

	receiverID, err := s.conversationRepo.FindPrivatePeerID(ctx, payload.ConversationID, input.SenderID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrConversationNotFound) {
			return &SendMessageResult{Failed: failedMessage(payload.ClientMsgID, formatID(payload.ConversationID), model.MessageSendStatusFailed, wsFailureConversationNotFound, "conversation not found")}, nil
		}
		return nil, apperrors.ErrInternal
	}

	existing, err := s.messageRepo.FindByClientMessageID(ctx, input.SenderID, payload.ConversationID, payload.ClientMsgID)
	if err != nil && !errors.Is(err, mysqlrepo.ErrMessageNotFound) {
		return nil, apperrors.ErrInternal
	}
	if existing != nil {
		return s.handleExistingMessage(existing, payload.MessageType, payload.Content, payload.ExtraJSON), nil
	}

	areFriends, err := s.friendRepo.AreFriends(ctx, input.SenderID, receiverID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if !areFriends {
		return &SendMessageResult{Failed: failedMessage(payload.ClientMsgID, formatID(payload.ConversationID), model.MessageSendStatusFailed, wsFailureNotFriends, "users are not friends")}, nil
	}

	blocked, err := s.friendRepo.IsBlocked(ctx, receiverID, input.SenderID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if blocked {
		return s.createBlockedMessage(ctx, input.SenderID, payload), nil
	}

	message := &model.Message{
		MessageID:      s.idGenerator.NextID(),
		ConversationID: payload.ConversationID,
		SenderID:       input.SenderID,
		ClientMsgID:    payload.ClientMsgID,
		MessageType:    payload.MessageType,
		Content:        sql.NullString{String: payload.Content, Valid: true},
		ExtraJSON:      sql.NullString{String: payload.ExtraJSON, Valid: true},
		SendStatus:     model.MessageSendStatusSent,
		CreatedAt:      now(),
	}

	if err := s.messageRepo.CreatePrivateMessage(ctx, message, receiverID); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateClientMessageID) {
			existing, findErr := s.messageRepo.FindByClientMessageID(ctx, input.SenderID, payload.ConversationID, payload.ClientMsgID)
			if findErr != nil {
				return nil, apperrors.ErrInternal
			}
			return s.handleExistingMessage(existing, payload.MessageType, payload.Content, payload.ExtraJSON), nil
		}
		return nil, apperrors.ErrInternal
	}

	return &SendMessageResult{
		Ack:        toAckOutput(message),
		Receive:    toReceiveOutput(message),
		ReceiverID: receiverID,
	}, nil
}

func (s *MessageService) ListConversationMessages(ctx context.Context, userID int64, conversationIDValue string, cursorValue string, limit int) (*MessagePageOutput, *apperrors.AppError) {
	conversationID, appErr := parsePositiveID(conversationIDValue)
	if appErr != nil {
		return nil, appErr
	}
	cursor := int64(0)
	if strings.TrimSpace(cursorValue) != "" {
		cursor, appErr = parsePositiveID(cursorValue)
		if appErr != nil {
			return nil, appErr
		}
	}
	limit = normalizeCursorLimit(limit)

	isMember, err := s.conversationRepo.IsActiveMember(ctx, conversationID, userID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if !isMember {
		return nil, apperrors.ErrMessageAccessDenied
	}

	messages, err := s.messageRepo.ListVisibleConversationMessages(ctx, userID, conversationID, cursor, limit)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	reverseMessages(messages)
	outputs := make([]MessageOutput, 0, len(messages))
	for i := range messages {
		outputs = append(outputs, toMessageOutput(&messages[i]))
	}

	nextCursor := ""
	if len(outputs) > 0 {
		nextCursor = outputs[0].MessageID
	}

	return &MessagePageOutput{
		List:       outputs,
		NextCursor: nextCursor,
		HasMore:    len(outputs) == limit,
		Limit:      limit,
	}, nil
}

func (s *MessageService) DeleteConversationMessage(ctx context.Context, userID int64, conversationIDValue string, messageIDValue string) *apperrors.AppError {
	conversationID, appErr := parsePositiveID(conversationIDValue)
	if appErr != nil {
		return appErr
	}
	messageID, appErr := parsePositiveID(messageIDValue)
	if appErr != nil {
		return appErr
	}

	if err := s.messageRepo.DeleteForUser(ctx, userID, conversationID, messageID, now()); err != nil {
		if errors.Is(err, mysqlrepo.ErrMessageNotFound) {
			return apperrors.ErrMessageNotFound
		}
		return apperrors.ErrInternal
	}
	return nil
}

func (s *MessageService) ClearConversationMessages(ctx context.Context, userID int64, conversationIDValue string) *apperrors.AppError {
	conversationID, appErr := parsePositiveID(conversationIDValue)
	if appErr != nil {
		return appErr
	}

	if err := s.conversationRepo.ClearMessagesForUser(ctx, conversationID, userID, sql.NullTime{Time: now(), Valid: true}); err != nil {
		if errors.Is(err, mysqlrepo.ErrConversationNotFound) {
			return apperrors.ErrMessageAccessDenied
		}
		return apperrors.ErrInternal
	}
	return nil
}

func (s *MessageService) RecallMessage(ctx context.Context, userID int64, messageIDValue string) (*RecallMessageOutput, *apperrors.AppError) {
	messageID, appErr := parsePositiveID(messageIDValue)
	if appErr != nil {
		return nil, appErr
	}
	if s.messageCacheRepo == nil {
		return nil, apperrors.ErrInternal
	}

	recalledAt := now()
	cacheWritten := false
	result, err := s.messageRepo.RecallMessage(ctx, mysqlrepo.RecallMessageParams{
		MessageID:    messageID,
		UserID:       userID,
		Now:          recalledAt,
		RecallWindow: s.recallWindow,
		CacheOriginalContent: func(content string) error {
			if err := s.messageCacheRepo.SaveRecallEditCache(ctx, messageID, userID, content, s.recallWindow); err != nil {
				return err
			}
			cacheWritten = true
			return nil
		},
	})
	if err != nil {
		if cacheWritten {
			_ = s.messageCacheRepo.DeleteRecallEditCache(context.Background(), messageID, userID)
		}
		return nil, mapMessageRepositoryError(err)
	}

	event := MessageRecalledEventOutput{
		MessageID:      formatID(result.MessageID),
		ConversationID: formatID(result.ConversationID),
		RecalledBy:     formatID(result.RecalledBy),
		RecalledAt:     formatTime(result.RecalledAt),
	}
	if s.recallNotifier != nil && len(result.RecipientIDs) > 0 {
		_ = s.recallNotifier.NotifyMessageRecalled(ctx, result.RecipientIDs, event)
	}

	return &RecallMessageOutput{
		MessageID:     formatID(result.MessageID),
		EditableUntil: formatTime(result.EditableUntil),
	}, nil
}

func (s *MessageService) GetRecallEditCache(ctx context.Context, userID int64, messageIDValue string) (*RecallEditCacheOutput, *apperrors.AppError) {
	messageID, appErr := parsePositiveID(messageIDValue)
	if appErr != nil {
		return nil, appErr
	}
	if s.messageCacheRepo == nil {
		return nil, apperrors.ErrInternal
	}

	message, err := s.messageRepo.FindByMessageID(ctx, messageID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrMessageNotFound) {
			return nil, apperrors.ErrMessageNotFound
		}
		return nil, apperrors.ErrInternal
	}
	if message.SenderID != userID {
		return nil, apperrors.ErrMessageAccessDenied
	}
	if !message.RecalledAt.Valid || !message.RecalledBy.Valid || message.RecalledBy.Int64 != userID {
		return nil, apperrors.ErrMessageNotRecallable
	}

	content, err := s.messageCacheRepo.GetRecallEditCache(ctx, messageID, userID)
	if err != nil {
		if errors.Is(err, redisrepo.ErrRecallEditCacheNotFound) {
			return nil, apperrors.ErrMessageNotRecallable
		}
		return nil, apperrors.ErrInternal
	}

	return &RecallEditCacheOutput{
		MessageID: formatID(messageID),
		Content:   content,
	}, nil
}

func (s *MessageService) validateSendInput(ctx context.Context, input SendMessageInput) (normalizedMessagePayload, *apperrors.AppError) {
	conversationID, appErr := parsePositiveID(input.ConversationID)
	if appErr != nil {
		return normalizedMessagePayload{}, appErr
	}

	clientMsgID := strings.TrimSpace(input.ClientMsgID)
	if clientMsgID == "" || utf8.RuneCountInString(clientMsgID) > 64 {
		return normalizedMessagePayload{}, apperrors.ErrInvalidParam
	}

	switch input.MessageType {
	case model.MessageTypeText:
		return s.validateTextMessageInput(input, conversationID, clientMsgID)
	case model.MessageTypeFile:
		return s.validateFileMessageInput(ctx, input, conversationID, clientMsgID)
	default:
		return normalizedMessagePayload{}, apperrors.ErrMessageInvalidContent
	}
}

func (s *MessageService) validateTextMessageInput(input SendMessageInput, conversationID int64, clientMsgID string) (normalizedMessagePayload, *apperrors.AppError) {
	content := strings.TrimSpace(input.Content)
	if content == "" || utf8.RuneCountInString(content) > s.textMessageMaxLength {
		return normalizedMessagePayload{}, apperrors.ErrMessageInvalidContent
	}

	extraJSON, appErr := normalizeExtraJSON(input.ExtraJSON)
	if appErr != nil {
		return normalizedMessagePayload{}, appErr
	}

	return normalizedMessagePayload{
		ConversationID: conversationID,
		ClientMsgID:    clientMsgID,
		MessageType:    model.MessageTypeText,
		Content:        content,
		ExtraJSON:      extraJSON,
	}, nil
}

func (s *MessageService) validateFileMessageInput(ctx context.Context, input SendMessageInput, conversationID int64, clientMsgID string) (normalizedMessagePayload, *apperrors.AppError) {
	if s.fileRepo == nil {
		return normalizedMessagePayload{}, apperrors.ErrInternal
	}

	fileID, appErr := parsePositiveID(input.Content)
	if appErr != nil {
		return normalizedMessagePayload{}, apperrors.ErrMessageInvalidContent
	}

	file, err := s.fileRepo.FindByFileID(ctx, fileID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrFileNotFound) {
			return normalizedMessagePayload{}, apperrors.ErrFileNotFound
		}
		return normalizedMessagePayload{}, apperrors.ErrInternal
	}
	if file.UploaderID != input.SenderID {
		return normalizedMessagePayload{}, apperrors.ErrFileAccessDenied
	}

	extraJSON, err := json.Marshal(fileMessageExtraOutput{
		FileID:   formatID(file.FileID),
		FileName: file.OriginalName,
		FileSize: file.FileSize,
		MimeType: file.MimeType.String,
	})
	if err != nil {
		return normalizedMessagePayload{}, apperrors.ErrInternal
	}

	return normalizedMessagePayload{
		ConversationID: conversationID,
		ClientMsgID:    clientMsgID,
		MessageType:    model.MessageTypeFile,
		Content:        formatID(file.FileID),
		ExtraJSON:      string(extraJSON),
	}, nil
}

type fileMessageExtraOutput struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

func (s *MessageService) handleExistingMessage(existing *model.Message, messageType string, content string, extraJSON string) *SendMessageResult {
	if existing.MessageType != messageType ||
		existing.Content.String != content ||
		existing.ExtraJSON.String != extraJSON {
		return &SendMessageResult{
			Failed: failedMessageWithServerMessage(
				existing,
				existing.ClientMsgID,
				formatID(existing.ConversationID),
				model.MessageSendStatusFailed,
				wsFailureDuplicateClientMsgIDConflict,
				"client_msg_id already exists with different content",
			),
		}
	}

	if existing.SendStatus == model.MessageSendStatusFailedBlocked {
		return &SendMessageResult{
			Failed:     failedMessageFromMessage(existing, wsFailureBlocked, "对方已拒收你的消息"),
			Duplicated: true,
		}
	}

	return &SendMessageResult{
		Ack:        toAckOutput(existing),
		Duplicated: true,
	}
}

func (s *MessageService) createBlockedMessage(ctx context.Context, senderID int64, payload normalizedMessagePayload) *SendMessageResult {
	message := &model.Message{
		MessageID:      s.idGenerator.NextID(),
		ConversationID: payload.ConversationID,
		SenderID:       senderID,
		ClientMsgID:    payload.ClientMsgID,
		MessageType:    payload.MessageType,
		Content:        sql.NullString{String: payload.Content, Valid: true},
		ExtraJSON:      sql.NullString{String: payload.ExtraJSON, Valid: true},
		SendStatus:     model.MessageSendStatusFailedBlocked,
		CreatedAt:      now(),
	}

	if err := s.messageRepo.CreateBlockedPrivateMessage(ctx, message); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateClientMessageID) {
			existing, findErr := s.messageRepo.FindByClientMessageID(ctx, senderID, payload.ConversationID, payload.ClientMsgID)
			if findErr != nil {
				return &SendMessageResult{Failed: failedMessage(payload.ClientMsgID, formatID(payload.ConversationID), model.MessageSendStatusFailed, wsFailureInternal, "message send failed")}
			}
			return s.handleExistingMessage(existing, payload.MessageType, payload.Content, payload.ExtraJSON)
		}
		return &SendMessageResult{Failed: failedMessage(payload.ClientMsgID, formatID(payload.ConversationID), model.MessageSendStatusFailed, wsFailureInternal, "message send failed")}
	}

	return &SendMessageResult{
		Failed: failedMessageFromMessage(message, wsFailureBlocked, "对方已拒收你的消息"),
	}
}

func normalizeExtraJSON(raw json.RawMessage) (string, *apperrors.AppError) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "{}", nil
	}

	var value any
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return "", apperrors.ErrInvalidParam
	}
	if value == nil {
		return "{}", nil
	}

	normalized, err := json.Marshal(value)
	if err != nil {
		return "", apperrors.ErrInvalidParam
	}
	return string(normalized), nil
}

func sendValidationFailureCode(appErr *apperrors.AppError) string {
	if appErr == nil {
		return wsFailureInvalidRequest
	}
	switch appErr.Code {
	case apperrors.ErrFileNotFound.Code:
		return wsFailureFileNotFound
	case apperrors.ErrFileAccessDenied.Code:
		return wsFailureFileAccessDenied
	default:
		return wsFailureInvalidRequest
	}
}

func failedMessage(clientMsgID string, conversationID string, sendStatus string, code string, message string) *SendMessageFailedOutput {
	return &SendMessageFailedOutput{
		ClientMsgID:    clientMsgID,
		ConversationID: conversationID,
		SendStatus:     sendStatus,
		Code:           code,
		Message:        message,
	}
}

func failedMessageFromMessage(message *model.Message, code string, text string) *SendMessageFailedOutput {
	return failedMessageWithServerMessage(message, message.ClientMsgID, formatID(message.ConversationID), message.SendStatus, code, text)
}

func failedMessageWithServerMessage(message *model.Message, clientMsgID string, conversationID string, sendStatus string, code string, text string) *SendMessageFailedOutput {
	output := failedMessage(clientMsgID, conversationID, sendStatus, code, text)
	if message != nil && message.MessageID > 0 {
		output.MessageID = formatID(message.MessageID)
		output.ServerTime = formatTime(message.CreatedAt)
	}
	return output
}

func toAckOutput(message *model.Message) *SendMessageAckOutput {
	return &SendMessageAckOutput{
		ClientMsgID:    message.ClientMsgID,
		MessageID:      formatID(message.MessageID),
		ConversationID: formatID(message.ConversationID),
		SendStatus:     message.SendStatus,
		ServerTime:     formatTime(message.CreatedAt),
	}
}

func toReceiveOutput(message *model.Message) *MessageReceiveOutput {
	return &MessageReceiveOutput{
		ClientMsgID:    message.ClientMsgID,
		MessageID:      formatID(message.MessageID),
		ConversationID: formatID(message.ConversationID),
		SenderID:       formatID(message.SenderID),
		MessageType:    message.MessageType,
		Content:        message.Content.String,
		ExtraJSON:      messageExtraJSON(message),
		CreatedAt:      formatTime(message.CreatedAt),
	}
}

func toMessageOutput(message *model.Message) MessageOutput {
	return MessageOutput{
		ClientMsgID:    message.ClientMsgID,
		MessageID:      formatID(message.MessageID),
		ConversationID: formatID(message.ConversationID),
		SenderID:       formatID(message.SenderID),
		MessageType:    message.MessageType,
		Content:        message.Content.String,
		ExtraJSON:      messageExtraJSON(message),
		SendStatus:     message.SendStatus,
		CreatedAt:      formatTime(message.CreatedAt),
	}
}

func messageExtraJSON(message *model.Message) json.RawMessage {
	if message.ExtraJSON.Valid && strings.TrimSpace(message.ExtraJSON.String) != "" {
		return json.RawMessage(message.ExtraJSON.String)
	}
	return json.RawMessage("{}")
}

func reverseMessages(messages []model.Message) {
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
}

func normalizeCursorLimit(limit int) int {
	if limit <= 0 {
		return 30
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func mapMessageRepositoryError(err error) *apperrors.AppError {
	switch {
	case errors.Is(err, mysqlrepo.ErrMessageNotFound):
		return apperrors.ErrMessageNotFound
	case errors.Is(err, mysqlrepo.ErrMessageAccessDenied):
		return apperrors.ErrMessageAccessDenied
	case errors.Is(err, mysqlrepo.ErrMessageNotRecallable):
		return apperrors.ErrMessageNotRecallable
	case errors.Is(err, mysqlrepo.ErrMessageAlreadyRecalled):
		return apperrors.ErrMessageRecalled
	default:
		return apperrors.ErrInternal
	}
}

func now() time.Time {
	return time.Now()
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05")
}
