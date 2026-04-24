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

type SearchUsersParams struct {
	Keyword  string
	ByUserID bool
	UserID   int64
	Limit    int
	Offset   int
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

func (r *UserRepository) Search(ctx context.Context, params SearchUsersParams) ([]model.UserWithProfile, int64, error) {
	where := "u.deleted_at IS NULL AND u.status = ?"
	args := []any{model.UserStatusNormal}
	if params.ByUserID {
		where += " AND u.user_id = ?"
		args = append(args, params.UserID)
	} else {
		where += " AND p.nickname LIKE ?"
		args = append(args, "%"+params.Keyword+"%")
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM users u
INNER JOIN user_profiles p ON p.user_id = u.user_id
WHERE ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.UserWithProfile{}, 0, nil
	}

	query := userWithProfileSelect + `
WHERE ` + where + `
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?
`
	queryArgs := append(args, params.Limit, params.Offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]model.UserWithProfile, 0)
	for rows.Next() {
		user, err := scanUserWithProfile(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
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
	row := r.db.QueryRowContext(ctx, userWithProfileSelect+`
WHERE `+where+` AND u.deleted_at IS NULL
LIMIT 1
`, arg)

	result, err := scanUserWithProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return result, nil
}

const userWithProfileSelect = `
SELECT
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM users u
INNER JOIN user_profiles p ON p.user_id = u.user_id
`

type scanner interface {
	Scan(dest ...any) error
}

func scanUserWithProfile(scanner scanner) (*model.UserWithProfile, error) {
	var result model.UserWithProfile
	if err := scanner.Scan(
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
