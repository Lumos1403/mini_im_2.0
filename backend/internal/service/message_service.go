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
)

const (
	wsFailureInvalidRequest               = "invalid_request"
	wsFailureConversationNotFound         = "conversation_not_found"
	wsFailureNotFriends                   = "not_friends"
	wsFailureBlocked                      = "failed_blocked"
	wsFailureDuplicateClientMsgIDConflict = "duplicate_client_msg_id_conflict"
	wsFailureInternal                     = "internal_error"
)

type SendTextMessageInput struct {
	SenderID       int64
	ConversationID string
	ClientMsgID    string
	MessageType    string
	Content        string
	ExtraJSON      json.RawMessage
}

type SendTextMessageResult struct {
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
	idGenerator          *snowflake.Node
	textMessageMaxLength int
}

func NewMessageService(
	conversationRepo *mysqlrepo.ConversationRepository,
	friendRepo *mysqlrepo.FriendRepository,
	messageRepo *mysqlrepo.MessageRepository,
	idGenerator *snowflake.Node,
	textMessageMaxLength int,
) *MessageService {
	if textMessageMaxLength <= 0 {
		textMessageMaxLength = 2000
	}
	return &MessageService{
		conversationRepo:     conversationRepo,
		friendRepo:           friendRepo,
		messageRepo:          messageRepo,
		idGenerator:          idGenerator,
		textMessageMaxLength: textMessageMaxLength,
	}
}

func (s *MessageService) SendTextMessage(ctx context.Context, input SendTextMessageInput) (*SendTextMessageResult, *apperrors.AppError) {
	conversationID, clientMsgID, content, extraJSON, appErr := s.validateSendInput(input)
	if appErr != nil {
		return &SendTextMessageResult{Failed: failedMessage(input.ClientMsgID, input.ConversationID, model.MessageSendStatusFailed, wsFailureInvalidRequest, appErr.Message)}, nil
	}

	receiverID, err := s.conversationRepo.FindPrivatePeerID(ctx, conversationID, input.SenderID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrConversationNotFound) {
			return &SendTextMessageResult{Failed: failedMessage(clientMsgID, formatID(conversationID), model.MessageSendStatusFailed, wsFailureConversationNotFound, "conversation not found")}, nil
		}
		return nil, apperrors.ErrInternal
	}

	existing, err := s.messageRepo.FindByClientMessageID(ctx, input.SenderID, conversationID, clientMsgID)
	if err != nil && !errors.Is(err, mysqlrepo.ErrMessageNotFound) {
		return nil, apperrors.ErrInternal
	}
	if existing != nil {
		return s.handleExistingMessage(existing, content, extraJSON), nil
	}

	areFriends, err := s.friendRepo.AreFriends(ctx, input.SenderID, receiverID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if !areFriends {
		return &SendTextMessageResult{Failed: failedMessage(clientMsgID, formatID(conversationID), model.MessageSendStatusFailed, wsFailureNotFriends, "users are not friends")}, nil
	}

	blocked, err := s.friendRepo.IsBlocked(ctx, receiverID, input.SenderID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if blocked {
		return s.createBlockedMessage(ctx, input.SenderID, conversationID, clientMsgID, content, extraJSON), nil
	}

	message := &model.Message{
		MessageID:      s.idGenerator.NextID(),
		ConversationID: conversationID,
		SenderID:       input.SenderID,
		ClientMsgID:    clientMsgID,
		MessageType:    model.MessageTypeText,
		Content:        sql.NullString{String: content, Valid: true},
		ExtraJSON:      sql.NullString{String: extraJSON, Valid: true},
		SendStatus:     model.MessageSendStatusSent,
		CreatedAt:      now(),
	}

	if err := s.messageRepo.CreatePrivateTextMessage(ctx, message, receiverID); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateClientMessageID) {
			existing, findErr := s.messageRepo.FindByClientMessageID(ctx, input.SenderID, conversationID, clientMsgID)
			if findErr != nil {
				return nil, apperrors.ErrInternal
			}
			return s.handleExistingMessage(existing, content, extraJSON), nil
		}
		return nil, apperrors.ErrInternal
	}

	return &SendTextMessageResult{
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

	if _, err := s.conversationRepo.FindPrivatePeerID(ctx, conversationID, userID); err != nil {
		if errors.Is(err, mysqlrepo.ErrConversationNotFound) {
			return nil, apperrors.ErrMessageAccessDenied
		}
		return nil, apperrors.ErrInternal
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

func (s *MessageService) validateSendInput(input SendTextMessageInput) (int64, string, string, string, *apperrors.AppError) {
	conversationID, appErr := parsePositiveID(input.ConversationID)
	if appErr != nil {
		return 0, "", "", "", appErr
	}

	clientMsgID := strings.TrimSpace(input.ClientMsgID)
	if clientMsgID == "" || utf8.RuneCountInString(clientMsgID) > 64 {
		return 0, "", "", "", apperrors.ErrInvalidParam
	}

	if input.MessageType != model.MessageTypeText {
		return 0, "", "", "", apperrors.ErrMessageInvalidContent
	}

	content := strings.TrimSpace(input.Content)
	if content == "" || utf8.RuneCountInString(content) > s.textMessageMaxLength {
		return 0, "", "", "", apperrors.ErrMessageInvalidContent
	}

	extraJSON, appErr := normalizeExtraJSON(input.ExtraJSON)
	if appErr != nil {
		return 0, "", "", "", appErr
	}

	return conversationID, clientMsgID, content, extraJSON, nil
}

func (s *MessageService) handleExistingMessage(existing *model.Message, content string, extraJSON string) *SendTextMessageResult {
	if existing.MessageType != model.MessageTypeText ||
		existing.Content.String != content ||
		existing.ExtraJSON.String != extraJSON {
		return &SendTextMessageResult{
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
		return &SendTextMessageResult{
			Failed:     failedMessageFromMessage(existing, wsFailureBlocked, "对方已拒收你的消息"),
			Duplicated: true,
		}
	}

	return &SendTextMessageResult{
		Ack:        toAckOutput(existing),
		Duplicated: true,
	}
}

func (s *MessageService) createBlockedMessage(ctx context.Context, senderID int64, conversationID int64, clientMsgID string, content string, extraJSON string) *SendTextMessageResult {
	message := &model.Message{
		MessageID:      s.idGenerator.NextID(),
		ConversationID: conversationID,
		SenderID:       senderID,
		ClientMsgID:    clientMsgID,
		MessageType:    model.MessageTypeText,
		Content:        sql.NullString{String: content, Valid: true},
		ExtraJSON:      sql.NullString{String: extraJSON, Valid: true},
		SendStatus:     model.MessageSendStatusFailedBlocked,
		CreatedAt:      now(),
	}

	if err := s.messageRepo.CreateBlockedPrivateTextMessage(ctx, message); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateClientMessageID) {
			existing, findErr := s.messageRepo.FindByClientMessageID(ctx, senderID, conversationID, clientMsgID)
			if findErr != nil {
				return &SendTextMessageResult{Failed: failedMessage(clientMsgID, formatID(conversationID), model.MessageSendStatusFailed, wsFailureInternal, "message send failed")}
			}
			return s.handleExistingMessage(existing, content, extraJSON)
		}
		return &SendTextMessageResult{Failed: failedMessage(clientMsgID, formatID(conversationID), model.MessageSendStatusFailed, wsFailureInternal, "message send failed")}
	}

	return &SendTextMessageResult{
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

func now() time.Time {
	return time.Now()
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05")
}
