package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mini_im/backend/internal/model"
)

var (
	ErrMessageNotFound          = errors.New("message not found")
	ErrMessageAccessDenied      = errors.New("message access denied")
	ErrMessageNotRecallable     = errors.New("message not recallable")
	ErrMessageAlreadyRecalled   = errors.New("message already recalled")
	ErrDuplicateClientMessageID = errors.New("duplicate client message id")
)

type MessageRepository struct {
	db *sql.DB
}

type RecallMessageParams struct {
	MessageID            int64
	UserID               int64
	Now                  time.Time
	RecallWindow         time.Duration
	CacheOriginalContent func(content string) error
}

type RecallMessageResult struct {
	MessageID      int64
	ConversationID int64
	RecalledBy     int64
	RecalledAt     time.Time
	EditableUntil  time.Time
	RecipientIDs   []int64
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) CreatePrivateMessage(ctx context.Context, message *model.Message, receiverID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := insertMessage(ctx, tx, message); err != nil {
		return err
	}

	if err := insertMessageUserState(ctx, tx, message.MessageID, message.SenderID); err != nil {
		return err
	}
	if err := insertMessageUserState(ctx, tx, message.MessageID, receiverID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE conversations
SET last_message_id = ?, last_message_at = ?
WHERE conversation_id = ?
  AND (last_message_id IS NULL OR last_message_id < ?)
`, message.MessageID, message.CreatedAt, message.ConversationID, message.MessageID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessageRepository) CreateGroupMessage(ctx context.Context, message *model.Message, groupID int64) ([]int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := insertMessage(ctx, tx, message); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT user_id
FROM group_members
WHERE group_id = ?
  AND status = ?
ORDER BY joined_at ASC
`, groupID, model.GroupMemberStatusActive)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			rows.Close()
			return nil, err
		}
		memberIDs = append(memberIDs, userID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, userID := range memberIDs {
		if err := insertMessageUserState(ctx, tx, message.MessageID, userID); err != nil {
			return nil, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE conversations
SET last_message_id = ?, last_message_at = ?
WHERE conversation_id = ?
  AND (last_message_id IS NULL OR last_message_id < ?)
`, message.MessageID, message.CreatedAt, message.ConversationID, message.MessageID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return memberIDs, nil
}

func (r *MessageRepository) CreateBlockedPrivateMessage(ctx context.Context, message *model.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := insertMessage(ctx, tx, message); err != nil {
		return err
	}
	if err := insertMessageUserState(ctx, tx, message.MessageID, message.SenderID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MessageRepository) FindByClientMessageID(ctx context.Context, senderID int64, conversationID int64, clientMsgID string) (*model.Message, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT
  id, message_id, conversation_id, sender_id, client_msg_id,
  message_type, content, extra_json, send_status, created_at, updated_at,
  recalled_at, recalled_by, is_deleted_all
FROM messages
WHERE sender_id = ? AND conversation_id = ? AND client_msg_id = ?
LIMIT 1
`, senderID, conversationID, clientMsgID)

	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) FindByMessageID(ctx context.Context, messageID int64) (*model.Message, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT
  id, message_id, conversation_id, sender_id, client_msg_id,
  message_type, content, extra_json, send_status, created_at, updated_at,
  recalled_at, recalled_by, is_deleted_all
FROM messages
WHERE message_id = ?
LIMIT 1
`, messageID)

	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) ListVisibleConversationMessages(ctx context.Context, userID int64, conversationID int64, cursor int64, limit int) ([]model.Message, error) {
	args := []any{userID, userID, userID, conversationID, model.MessageSendStatusSent, model.MessageSendStatusFailedBlocked}
	cursorClause := ""
	if cursor > 0 {
		cursorClause = "  AND m.message_id < ?\n"
		args = append(args, cursor)
	}
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, `
SELECT
  m.id, m.message_id, m.conversation_id, m.sender_id, m.client_msg_id,
  m.message_type, m.content, m.extra_json, m.send_status, m.created_at, m.updated_at,
  m.recalled_at, m.recalled_by, m.is_deleted_all
FROM messages m
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = 'active'
INNER JOIN conversation_user_states cus
  ON cus.conversation_id = m.conversation_id
  AND cus.user_id = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
  AND mus.is_deleted = 0
WHERE m.conversation_id = ?
  AND m.send_status IN (?, ?)
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
`+cursorClause+`ORDER BY m.message_id DESC
LIMIT ?
`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *MessageRepository) DeleteForUser(ctx context.Context, userID int64, conversationID int64, messageID int64, deletedAt time.Time) error {
	exists, err := r.userMessageStateExists(ctx, userID, conversationID, messageID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMessageNotFound
	}

	_, err = r.db.ExecContext(ctx, `
UPDATE message_user_states
SET is_deleted = 1,
    deleted_at = COALESCE(deleted_at, ?),
    updated_at = CURRENT_TIMESTAMP
WHERE message_id = ?
  AND user_id = ?
`, deletedAt, messageID, userID)
	return err
}

func (r *MessageRepository) RecallMessage(ctx context.Context, params RecallMessageParams) (*RecallMessageResult, error) {
	if params.MessageID <= 0 || params.UserID <= 0 || params.RecallWindow <= 0 || params.CacheOriginalContent == nil {
		return nil, ErrMessageNotRecallable
	}
	if params.Now.IsZero() {
		params.Now = time.Now()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	message, err := findMessageByIDForUpdate(ctx, tx, params.MessageID)
	if err != nil {
		return nil, err
	}
	if message.IsDeletedAll {
		return nil, ErrMessageNotFound
	}

	isMember, err := isActiveConversationMember(ctx, tx, message.ConversationID, params.UserID)
	if err != nil {
		return nil, err
	}
	if !isMember || message.SenderID != params.UserID {
		return nil, ErrMessageAccessDenied
	}

	if message.RecalledAt.Valid || message.SendStatus == model.MessageSendStatusRecalled {
		return nil, ErrMessageAlreadyRecalled
	}
	if message.SendStatus != model.MessageSendStatusSent {
		return nil, ErrMessageNotRecallable
	}

	editableUntil := params.Now.Add(params.RecallWindow)
	if params.Now.After(message.CreatedAt.Add(params.RecallWindow)) {
		return nil, ErrMessageNotRecallable
	}

	originalContent := ""
	if message.Content.Valid {
		originalContent = message.Content.String
	}
	if err := params.CacheOriginalContent(originalContent); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
UPDATE messages
SET content = NULL,
    send_status = ?,
    recalled_at = ?,
    recalled_by = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE message_id = ?
  AND recalled_at IS NULL
`, model.MessageSendStatusRecalled, params.Now, params.UserID, message.MessageID)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrMessageAlreadyRecalled
	}

	if err := rollbackConversationLastMessage(ctx, tx, message.ConversationID, message.MessageID); err != nil {
		return nil, err
	}

	recipientIDs, err := listVisibleMessageUserIDs(ctx, tx, message.MessageID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &RecallMessageResult{
		MessageID:      message.MessageID,
		ConversationID: message.ConversationID,
		RecalledBy:     params.UserID,
		RecalledAt:     params.Now,
		EditableUntil:  editableUntil,
		RecipientIDs:   recipientIDs,
	}, nil
}

func (r *MessageRepository) userMessageStateExists(ctx context.Context, userID int64, conversationID int64, messageID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
SELECT 1
FROM messages m
INNER JOIN conversations c
  ON c.conversation_id = m.conversation_id
  AND c.status = ?
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
WHERE m.conversation_id = ?
  AND m.message_id = ?
LIMIT 1
`, model.ConversationStatusNormal, userID, model.ConversationMemberStatusActive, userID, conversationID, messageID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func insertMessageUserState(ctx context.Context, exec Executor, messageID int64, userID int64) error {
	_, err := exec.ExecContext(ctx, `
INSERT INTO message_user_states (message_id, user_id, is_deleted)
VALUES (?, ?, 0)
`, messageID, userID)
	return err
}

func insertMessage(ctx context.Context, exec Executor, message *model.Message) error {
	_, err := exec.ExecContext(ctx, `
INSERT INTO messages (
  message_id, conversation_id, sender_id, client_msg_id,
  message_type, content, extra_json, send_status, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		message.MessageID,
		message.ConversationID,
		message.SenderID,
		message.ClientMsgID,
		message.MessageType,
		message.Content,
		message.ExtraJSON.String,
		message.SendStatus,
		message.CreatedAt,
	)
	if isDuplicateEntry(err) {
		return ErrDuplicateClientMessageID
	}
	return err
}

func findMessageByIDForUpdate(ctx context.Context, exec Executor, messageID int64) (*model.Message, error) {
	row := exec.QueryRowContext(ctx, `
SELECT
  id, message_id, conversation_id, sender_id, client_msg_id,
  message_type, content, extra_json, send_status, created_at, updated_at,
  recalled_at, recalled_by, is_deleted_all
FROM messages
WHERE message_id = ?
LIMIT 1
FOR UPDATE
`, messageID)

	message, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return message, nil
}

func isActiveConversationMember(ctx context.Context, exec Executor, conversationID int64, userID int64) (bool, error) {
	var exists int
	err := exec.QueryRowContext(ctx, `
SELECT 1
FROM conversations c
INNER JOIN conversation_members cm
  ON cm.conversation_id = c.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
WHERE c.conversation_id = ?
  AND c.status = ?
LIMIT 1
`, userID, model.ConversationMemberStatusActive, conversationID, model.ConversationStatusNormal).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func rollbackConversationLastMessage(ctx context.Context, exec Executor, conversationID int64, recalledMessageID int64) error {
	var lastMessageID sql.NullInt64
	if err := exec.QueryRowContext(ctx, `
SELECT last_message_id
FROM conversations
WHERE conversation_id = ?
FOR UPDATE
`, conversationID).Scan(&lastMessageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if !lastMessageID.Valid || lastMessageID.Int64 != recalledMessageID {
		return nil
	}

	var fallbackID int64
	var fallbackAt time.Time
	err := exec.QueryRowContext(ctx, `
SELECT message_id, created_at
FROM messages
WHERE conversation_id = ?
  AND message_id <> ?
  AND recalled_at IS NULL
  AND is_deleted_all = 0
  AND send_status = ?
ORDER BY message_id DESC
LIMIT 1
`, conversationID, recalledMessageID, model.MessageSendStatusSent).Scan(&fallbackID, &fallbackAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, updateErr := exec.ExecContext(ctx, `
UPDATE conversations
SET last_message_id = NULL,
    last_message_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ?
`, conversationID)
			return updateErr
		}
		return err
	}

	_, err = exec.ExecContext(ctx, `
UPDATE conversations
SET last_message_id = ?,
    last_message_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ?
`, fallbackID, fallbackAt, conversationID)
	return err
}

func listVisibleMessageUserIDs(ctx context.Context, exec Executor, messageID int64) ([]int64, error) {
	rows, err := exec.QueryContext(ctx, `
SELECT user_id
FROM message_user_states
WHERE message_id = ?
  AND is_deleted = 0
`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func scanMessage(row scanner) (*model.Message, error) {
	var message model.Message
	if err := row.Scan(
		&message.ID,
		&message.MessageID,
		&message.ConversationID,
		&message.SenderID,
		&message.ClientMsgID,
		&message.MessageType,
		&message.Content,
		&message.ExtraJSON,
		&message.SendStatus,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.RecalledAt,
		&message.RecalledBy,
		&message.IsDeletedAll,
	); err != nil {
		return nil, err
	}
	return &message, nil
}
