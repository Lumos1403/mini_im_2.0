package mysql

import (
	"context"
	"database/sql"
	"errors"

	"mini_im/backend/internal/model"

	drivermysql "github.com/go-sql-driver/mysql"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrDuplicateUser = errors.New("duplicate user")
)

type UserRepository struct {
	db *sql.DB
}

type UpdateProfileParams struct {
	Nickname  string
	AvatarURL string
	Gender    string
	Bio       string
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUserWithProfile(ctx context.Context, user *model.User, profile *model.UserProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO users (user_id, username, password_hash, user_type, status)
VALUES (?, ?, ?, ?, ?)
`, user.UserID, user.Username, user.PasswordHash, user.UserType, user.Status)
	if err != nil {
		return mapDuplicateError(err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO user_profiles (user_id, nickname, avatar_url, gender, bio, profile_status, profile_review_reason)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, profile.UserID, profile.Nickname, profile.AvatarURL, profile.Gender, profile.Bio, profile.ProfileStatus, profile.ProfileReviewReason)
	if err != nil {
		return mapDuplicateError(err)
	}

	return tx.Commit()
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.UserWithProfile, error) {
	return r.findOne(ctx, "u.username = ?", username)
}

func (r *UserRepository) FindByUserID(ctx context.Context, userID int64) (*model.UserWithProfile, error) {
	return r.findOne(ctx, "u.user_id = ?", userID)
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, profile UpdateProfileParams) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE user_profiles
SET nickname = ?, avatar_url = ?, gender = ?, bio = ?
WHERE user_id = ?
`, profile.Nickname, nullableString(profile.AvatarURL), nullableString(profile.Gender), nullableString(profile.Bio), userID)
	return err
}

func (r *UserRepository) findOne(ctx context.Context, where string, arg interface{}) (*model.UserWithProfile, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM users u
INNER JOIN user_profiles p ON p.user_id = u.user_id
WHERE `+where+` AND u.deleted_at IS NULL
LIMIT 1
`, arg)

	var result model.UserWithProfile
	if err := row.Scan(
		&result.User.ID,
		&result.User.UserID,
		&result.User.Username,
		&result.User.PasswordHash,
		&result.User.UserType,
		&result.User.Status,
		&result.User.CreatedAt,
		&result.User.UpdatedAt,
		&result.User.DeletedAt,
		&result.Profile.ID,
		&result.Profile.UserID,
		&result.Profile.Nickname,
		&result.Profile.AvatarURL,
		&result.Profile.Gender,
		&result.Profile.Bio,
		&result.Profile.ProfileStatus,
		&result.Profile.ProfileReviewReason,
		&result.Profile.CreatedAt,
		&result.Profile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &result, nil
}

func nullableString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func mapDuplicateError(err error) error {
	var mysqlErr *drivermysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicateUser
	}
	return err
}
