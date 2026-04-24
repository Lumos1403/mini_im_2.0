package mysql

import (
	"context"
	"database/sql"
	"errors"

	"mini_im/backend/internal/model"
)

var (
	ErrMessageNotFound          = errors.New("message not found")
	ErrDuplicateClientMessageID = errors.New("duplicate client message id")
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) CreatePrivateTextMessage(ctx context.Context, message *model.Message, receiverID int64) error {
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

func (r *MessageRepository) CreateBlockedPrivateTextMessage(ctx context.Context, message *model.Message) error {
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
