package model

import (
	"database/sql"
	"time"
)

type File struct {
	ID           int64
	FileID       int64
	UploaderID   int64
	OriginalName string
	StoragePath  string
	FileSize     int64
	MimeType     sql.NullString
	SHA256       sql.NullString
	CreatedAt    time.Time
}
