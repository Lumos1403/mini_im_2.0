package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
	"mini_im/backend/internal/storage"
)

const defaultMaxFileSizeMB = 50

type FileService struct {
	fileRepo     *mysqlrepo.FileRepository
	fileStorage  storage.FileStorage
	idGenerator  *snowflake.Node
	maxSizeBytes int64
}

type FileDownloadOutput struct {
	FileName string
	FileSize int64
	MimeType string
	Reader   io.ReadCloser
}

func NewFileService(fileRepo *mysqlrepo.FileRepository, fileStorage storage.FileStorage, idGenerator *snowflake.Node, maxSizeMB int) *FileService {
	if maxSizeMB <= 0 {
		maxSizeMB = defaultMaxFileSizeMB
	}
	return &FileService{
		fileRepo:     fileRepo,
		fileStorage:  fileStorage,
		idGenerator:  idGenerator,
		maxSizeBytes: int64(maxSizeMB) * 1024 * 1024,
	}
}

func (s *FileService) MaxUploadBytes() int64 {
	return s.maxSizeBytes
}

func (s *FileService) Upload(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (*FileUploadOutput, *apperrors.AppError) {
	if s == nil || s.fileRepo == nil || s.fileStorage == nil || s.idGenerator == nil {
		return nil, apperrors.ErrInternal
	}
	if userID <= 0 || fileHeader == nil {
		return nil, apperrors.ErrInvalidParam
	}
	if fileHeader.Size > s.maxSizeBytes {
		return nil, apperrors.ErrFileTooLarge
	}

	originalName, appErr := normalizeOriginalFileName(fileHeader.Filename)
	if appErr != nil {
		return nil, appErr
	}
	mimeType := normalizeMimeType(fileHeader.Header.Get("Content-Type"))

	src, err := fileHeader.Open()
	if err != nil {
		return nil, apperrors.ErrFileInvalid
	}
	defer src.Close()

	fileID := s.idGenerator.NextID()
	storagePath := buildStoragePath(fileID)
	stored, err := s.fileStorage.Save(ctx, src, storagePath)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if stored.Size > s.maxSizeBytes {
		_ = s.fileStorage.Delete(context.Background(), stored.StoragePath)
		return nil, apperrors.ErrFileTooLarge
	}

	file := &model.File{
		FileID:       fileID,
		UploaderID:   userID,
		OriginalName: originalName,
		StoragePath:  stored.StoragePath,
		FileSize:     stored.Size,
		MimeType:     nullableFileString(mimeType),
		SHA256:       nullableFileString(stored.SHA256),
		CreatedAt:    now(),
	}
	if err := s.fileRepo.Create(ctx, file); err != nil {
		_ = s.fileStorage.Delete(context.Background(), stored.StoragePath)
		return nil, apperrors.ErrInternal
	}

	return toFileUploadOutput(file), nil
}

func (s *FileService) PrepareDownload(ctx context.Context, userID int64, fileIDValue string) (*FileDownloadOutput, *apperrors.AppError) {
	if s == nil || s.fileRepo == nil || s.fileStorage == nil {
		return nil, apperrors.ErrInternal
	}

	fileID, appErr := parsePositiveID(fileIDValue)
	if appErr != nil {
		return nil, appErr
	}

	file, err := s.fileRepo.FindDownloadableByFileID(ctx, userID, fileID)
	if err != nil {
		return nil, mapFileRepositoryError(err)
	}

	reader, err := s.fileStorage.Open(ctx, file.StoragePath)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	return &FileDownloadOutput{
		FileName: file.OriginalName,
		FileSize: file.FileSize,
		MimeType: file.MimeType.String,
		Reader:   reader,
	}, nil
}

func normalizeOriginalFileName(filename string) (string, *apperrors.AppError) {
	filename = strings.TrimSpace(filename)
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Base(filename)
	filename = strings.TrimSpace(filename)
	if filename == "" || filename == "." || filename == "/" || utf8.RuneCountInString(filename) > 255 {
		return "", apperrors.ErrFileInvalid
	}
	return filename, nil
}

func normalizeMimeType(mimeType string) string {
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" || utf8.RuneCountInString(mimeType) > 128 {
		return "application/octet-stream"
	}
	return mimeType
}

func buildStoragePath(fileID int64) string {
	current := now()
	return filepath.ToSlash(filepath.Join(
		current.Format("2006"),
		current.Format("01"),
		current.Format("02"),
		strconv.FormatInt(fileID, 10),
	))
}

func toFileUploadOutput(file *model.File) *FileUploadOutput {
	return &FileUploadOutput{
		FileID:       formatID(file.FileID),
		OriginalName: file.OriginalName,
		FileSize:     file.FileSize,
		MimeType:     file.MimeType.String,
	}
}

func nullableFileString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func mapFileRepositoryError(err error) *apperrors.AppError {
	switch {
	case errors.Is(err, mysqlrepo.ErrFileNotFound):
		return apperrors.ErrFileNotFound
	case errors.Is(err, mysqlrepo.ErrFileAccessDenied):
		return apperrors.ErrFileAccessDenied
	default:
		return apperrors.ErrInternal
	}
}
