package mysql

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"mini_im/backend/internal/model"
)

var (
	ErrFileNotFound     = errors.New("file not found")
	ErrFileAccessDenied = errors.New("file access denied")
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO files (
  file_id, uploader_id, original_name, storage_path,
  file_size, mime_type, sha256, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, file.FileID, file.UploaderID, file.OriginalName, file.StoragePath, file.FileSize, file.MimeType, file.SHA256, file.CreatedAt)
	return err
}

func (r *FileRepository) FindByFileID(ctx context.Context, fileID int64) (*model.File, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, file_id, uploader_id, original_name, storage_path, file_size, mime_type, sha256, created_at
FROM files
WHERE file_id = ?
LIMIT 1
`, fileID)

	file, err := scanFile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return file, nil
}

func (r *FileRepository) FindDownloadableByFileID(ctx context.Context, userID int64, fileID int64) (*model.File, error) {
	file, err := r.FindByFileID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file.UploaderID == userID {
		return file, nil
	}

	allowed, err := r.hasVisibleFileMessage(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrFileAccessDenied
	}
	return file, nil
}

func (r *FileRepository) hasVisibleFileMessage(ctx context.Context, userID int64, fileID int64) (bool, error) {
	fileIDValue := strconv.FormatInt(fileID, 10)
	var exists int
	err := r.db.QueryRowContext(ctx, `
SELECT 1
FROM messages m
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
WHERE m.message_type = ?
  AND m.send_status = ?
  AND m.recalled_at IS NULL
  AND m.is_deleted_all = 0
  AND (cus.cleared_at IS NULL OR m.created_at > cus.cleared_at)
  AND (
    m.content = ?
    OR JSON_UNQUOTE(JSON_EXTRACT(m.extra_json, '$.file_id')) = ?
  )
LIMIT 1
`, userID, model.ConversationMemberStatusActive, userID, userID, model.MessageTypeFile, model.MessageSendStatusSent, fileIDValue, fileIDValue).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanFile(row scanner) (*model.File, error) {
	var file model.File
	if err := row.Scan(
		&file.ID,
		&file.FileID,
		&file.UploaderID,
		&file.OriginalName,
		&file.StoragePath,
		&file.FileSize,
		&file.MimeType,
		&file.SHA256,
		&file.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &file, nil
}
