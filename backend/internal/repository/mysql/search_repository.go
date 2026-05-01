package mysql

import (
	"context"
	"database/sql"
	"time"

	"mini_im/backend/internal/model"
)

type SearchRepository struct {
	db *sql.DB
}

func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

type MessageSearchRow struct {
	MessageID        int64
	ConversationID   int64
	ConversationType string
	SenderID         int64
	MessageType      string
	Content          sql.NullString
	CreatedAt        time.Time
}

type FileSearchRow struct {
	FileID           int64
	UploaderID       int64
	OriginalName     string
	FileSize         int64
	MimeType         sql.NullString
	MessageID        int64
	ConversationID   int64
	ConversationType string
	CreatedAt        time.Time
}

func (r *SearchRepository) SearchMessages(ctx context.Context, userID int64, keyword string, limit int, offset int) ([]MessageSearchRow, int64, error) {
	likeKeyword := "%" + keyword + "%"

	baseArgs := []any{
		model.ConversationStatusNormal,
		userID, model.ConversationMemberStatusActive,
		userID,
		userID,
		model.ConversationTypeGroup,
		userID,
		model.MessageSendStatusSent,
		model.MessageSendStatusFailedBlocked,
		model.MessageSendStatusFailedBlocked,
		userID,
		model.ConversationTypeGroup,
		model.GroupStatusNormal,
		model.GroupMemberStatusActive,
		likeKeyword,
	}

	// COUNT
	var total int64
	countArgs := make([]any, len(baseArgs))
	copy(countArgs, baseArgs)
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM messages m
INNER JOIN conversations c
  ON c.conversation_id = m.conversation_id
  AND c.status = ?
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
INNER JOIN conversation_user_states cus
  ON cus.conversation_id = m.conversation_id
  AND cus.user_id = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
  AND mus.is_deleted = 0
LEFT JOIN `+"`groups`"+` g
  ON g.conversation_id = m.conversation_id
  AND c.conversation_type = ?
LEFT JOIN group_members gm
  ON gm.group_id = g.group_id
  AND gm.user_id = ?
WHERE m.send_status IN (?, ?)
  AND (m.send_status <> ? OR m.sender_id = ?)
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
  AND (
    c.conversation_type <> ?
    OR (
      g.status = ?
      AND gm.status = ?
      AND m.created_at > gm.joined_at
    )
  )
  AND m.content LIKE ?
`, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []MessageSearchRow{}, 0, nil
	}

	// LIST
	listArgs := make([]any, len(baseArgs))
	copy(listArgs, baseArgs)
	listArgs = append(listArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, `
SELECT
  m.message_id, m.conversation_id, c.conversation_type,
  m.sender_id, m.message_type, m.content, m.created_at
FROM messages m
INNER JOIN conversations c
  ON c.conversation_id = m.conversation_id
  AND c.status = ?
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
INNER JOIN conversation_user_states cus
  ON cus.conversation_id = m.conversation_id
  AND cus.user_id = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
  AND mus.is_deleted = 0
LEFT JOIN `+"`groups`"+` g
  ON g.conversation_id = m.conversation_id
  AND c.conversation_type = ?
LEFT JOIN group_members gm
  ON gm.group_id = g.group_id
  AND gm.user_id = ?
WHERE m.send_status IN (?, ?)
  AND (m.send_status <> ? OR m.sender_id = ?)
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
  AND (
    c.conversation_type <> ?
    OR (
      g.status = ?
      AND gm.status = ?
      AND m.created_at > gm.joined_at
    )
  )
  AND m.content LIKE ?
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?
`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]MessageSearchRow, 0)
	for rows.Next() {
		var row MessageSearchRow
		if err := rows.Scan(
			&row.MessageID, &row.ConversationID, &row.ConversationType,
			&row.SenderID, &row.MessageType, &row.Content, &row.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *SearchRepository) SearchFiles(ctx context.Context, userID int64, keyword string, limit int, offset int) ([]FileSearchRow, int64, error) {
	likeKeyword := "%" + keyword + "%"

	baseArgs := []any{
		model.MessageTypeFile,
		model.ConversationStatusNormal,
		userID, model.ConversationMemberStatusActive,
		userID,
		userID,
		model.ConversationTypeGroup,
		userID,
		model.MessageSendStatusSent,
		model.MessageSendStatusFailedBlocked,
		model.MessageSendStatusFailedBlocked,
		userID,
		model.ConversationTypeGroup,
		model.GroupStatusNormal,
		model.GroupMemberStatusActive,
		likeKeyword,
	}

	// COUNT
	var total int64
	countArgs := make([]any, len(baseArgs))
	copy(countArgs, baseArgs)
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM files f
INNER JOIN messages m
  ON m.content = CONCAT('', f.file_id)
  AND m.message_type = ?
INNER JOIN conversations c
  ON c.conversation_id = m.conversation_id
  AND c.status = ?
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
INNER JOIN conversation_user_states cus
  ON cus.conversation_id = m.conversation_id
  AND cus.user_id = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
  AND mus.is_deleted = 0
LEFT JOIN `+"`groups`"+` g
  ON g.conversation_id = m.conversation_id
  AND c.conversation_type = ?
LEFT JOIN group_members gm
  ON gm.group_id = g.group_id
  AND gm.user_id = ?
WHERE m.send_status IN (?, ?)
  AND (m.send_status <> ? OR m.sender_id = ?)
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
  AND (
    c.conversation_type <> ?
    OR (
      g.status = ?
      AND gm.status = ?
      AND m.created_at > gm.joined_at
    )
  )
  AND f.original_name LIKE ?
`, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []FileSearchRow{}, 0, nil
	}

	// LIST
	listArgs := make([]any, len(baseArgs))
	copy(listArgs, baseArgs)
	listArgs = append(listArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, `
SELECT
  f.file_id, f.uploader_id, f.original_name, f.file_size, f.mime_type,
  m.message_id, m.conversation_id, c.conversation_type, m.created_at
FROM files f
INNER JOIN messages m
  ON m.content = CONCAT('', f.file_id)
  AND m.message_type = ?
INNER JOIN conversations c
  ON c.conversation_id = m.conversation_id
  AND c.status = ?
INNER JOIN conversation_members cm
  ON cm.conversation_id = m.conversation_id
  AND cm.user_id = ?
  AND cm.status = ?
INNER JOIN conversation_user_states cus
  ON cus.conversation_id = m.conversation_id
  AND cus.user_id = ?
INNER JOIN message_user_states mus
  ON mus.message_id = m.message_id
  AND mus.user_id = ?
  AND mus.is_deleted = 0
LEFT JOIN `+"`groups`"+` g
  ON g.conversation_id = m.conversation_id
  AND c.conversation_type = ?
LEFT JOIN group_members gm
  ON gm.group_id = g.group_id
  AND gm.user_id = ?
WHERE m.send_status IN (?, ?)
  AND (m.send_status <> ? OR m.sender_id = ?)
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
  AND (
    c.conversation_type <> ?
    OR (
      g.status = ?
      AND gm.status = ?
      AND m.created_at > gm.joined_at
    )
  )
  AND f.original_name LIKE ?
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?
`, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]FileSearchRow, 0)
	for rows.Next() {
		var row FileSearchRow
		if err := rows.Scan(
			&row.FileID, &row.UploaderID, &row.OriginalName, &row.FileSize, &row.MimeType,
			&row.MessageID, &row.ConversationID, &row.ConversationType, &row.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
