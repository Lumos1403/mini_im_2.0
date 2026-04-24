package mysql

import (
	"context"
	"database/sql"
	"errors"

	"mini_im/backend/internal/model"

	drivermysql "github.com/go-sql-driver/mysql"
)

var (
	ErrFriendRequestNotFound = errors.New("friend request not found")
	ErrFriendRequestPending  = errors.New("friend request pending")
	ErrFriendshipNotFound    = errors.New("friendship not found")
)

type FriendRepository struct {
	db *sql.DB
}

func NewFriendRepository(db *sql.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

func (r *FriendRepository) CreateFriendRequest(ctx context.Context, request *model.FriendRequest) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO friend_requests (request_id, from_user_id, to_user_id, message, status)
VALUES (?, ?, ?, ?, ?)
`, request.RequestID, request.FromUserID, request.ToUserID, request.Message, request.Status)
	if isDuplicateEntry(err) {
		return ErrFriendRequestPending
	}
	return err
}

func (r *FriendRepository) HasPendingRequestBetween(ctx context.Context, userID1 int64, userID2 int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
SELECT 1
FROM friend_requests
WHERE status = ?
  AND (
    (from_user_id = ? AND to_user_id = ?)
    OR (from_user_id = ? AND to_user_id = ?)
  )
LIMIT 1
`, model.FriendRequestStatusPending, userID1, userID2, userID2, userID1).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *FriendRepository) AreFriends(ctx context.Context, userID1 int64, userID2 int64) (bool, error) {
	left, right := orderedFriendPair(userID1, userID2)
	if left == right {
		return false, nil
	}

	var status string
	err := r.db.QueryRowContext(ctx, `
SELECT status
FROM friendships
WHERE user_id_1 = ? AND user_id_2 = ?
LIMIT 1
`, left, right).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == model.FriendshipStatusNormal, nil
}

func (r *FriendRepository) ListFriendRequests(ctx context.Context, userID int64, received bool, limit int, offset int) ([]model.FriendRequestWithUser, int64, error) {
	where := "fr.to_user_id = ?"
	peerID := "fr.from_user_id"
	if !received {
		where = "fr.from_user_id = ?"
		peerID = "fr.to_user_id"
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM friend_requests fr
WHERE ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.FriendRequestWithUser{}, 0, nil
	}

	query := `
SELECT
  fr.id, fr.request_id, fr.from_user_id, fr.to_user_id, fr.message, fr.status, fr.created_at, fr.updated_at,
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM friend_requests fr
INNER JOIN users u ON u.user_id = ` + peerID + `
INNER JOIN user_profiles p ON p.user_id = u.user_id
WHERE ` + where + `
ORDER BY fr.created_at DESC
LIMIT ? OFFSET ?
`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	requests := make([]model.FriendRequestWithUser, 0)
	for rows.Next() {
		item, err := scanFriendRequestWithUser(rows)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *FriendRepository) AcceptFriendRequest(ctx context.Context, requestID int64, receiverID int64, userID1 int64, userID2 int64) error {
	left, right := orderedFriendPair(userID1, userID2)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
UPDATE friend_requests
SET status = ?
WHERE request_id = ? AND to_user_id = ? AND status = ?
`, model.FriendRequestStatusAccepted, requestID, receiverID, model.FriendRequestStatusPending)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrFriendRequestNotFound
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO friendships (user_id_1, user_id_2, status)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE status = VALUES(status), updated_at = CURRENT_TIMESTAMP
`, left, right, model.FriendshipStatusNormal)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *FriendRepository) RejectFriendRequest(ctx context.Context, requestID int64, receiverID int64) error {
	result, err := r.db.ExecContext(ctx, `
UPDATE friend_requests
SET status = ?
WHERE request_id = ? AND to_user_id = ? AND status = ?
`, model.FriendRequestStatusRejected, requestID, receiverID, model.FriendRequestStatusPending)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrFriendRequestNotFound
	}
	return nil
}

func (r *FriendRepository) FindFriendRequest(ctx context.Context, requestID int64) (*model.FriendRequest, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, request_id, from_user_id, to_user_id, message, status, created_at, updated_at
FROM friend_requests
WHERE request_id = ?
LIMIT 1
`, requestID)

	var request model.FriendRequest
	if err := row.Scan(
		&request.ID,
		&request.RequestID,
		&request.FromUserID,
		&request.ToUserID,
		&request.Message,
		&request.Status,
		&request.CreatedAt,
		&request.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFriendRequestNotFound
		}
		return nil, err
	}

	return &request, nil
}

func (r *FriendRepository) ListFriends(ctx context.Context, userID int64, limit int, offset int) ([]model.FriendshipWithUser, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM friendships f
INNER JOIN users u ON u.user_id = CASE WHEN f.user_id_1 = ? THEN f.user_id_2 ELSE f.user_id_1 END
WHERE (f.user_id_1 = ? OR f.user_id_2 = ?)
  AND f.status = ?
  AND u.deleted_at IS NULL
`, userID, userID, userID, model.FriendshipStatusNormal).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.FriendshipWithUser{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  f.id, f.user_id_1, f.user_id_2, f.status, f.created_at, f.updated_at,
  u.id, u.user_id, u.username, u.password_hash, u.user_type, u.status, u.created_at, u.updated_at, u.deleted_at,
  p.id, p.user_id, p.nickname, p.avatar_url, p.gender, p.bio, p.profile_status, p.profile_review_reason, p.created_at, p.updated_at
FROM friendships f
INNER JOIN users u ON u.user_id = CASE WHEN f.user_id_1 = ? THEN f.user_id_2 ELSE f.user_id_1 END
INNER JOIN user_profiles p ON p.user_id = u.user_id
WHERE (f.user_id_1 = ? OR f.user_id_2 = ?)
  AND f.status = ?
  AND u.deleted_at IS NULL
ORDER BY f.updated_at DESC
LIMIT ? OFFSET ?
`, userID, userID, userID, model.FriendshipStatusNormal, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	friends := make([]model.FriendshipWithUser, 0)
	for rows.Next() {
		item, err := scanFriendshipWithUser(rows)
		if err != nil {
			return nil, 0, err
		}
		friends = append(friends, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return friends, total, nil
}

func (r *FriendRepository) DeleteFriend(ctx context.Context, userID1 int64, userID2 int64) error {
	left, right := orderedFriendPair(userID1, userID2)
	result, err := r.db.ExecContext(ctx, `
UPDATE friendships
SET status = ?
WHERE user_id_1 = ? AND user_id_2 = ? AND status = ?
`, model.FriendshipStatusDeleted, left, right, model.FriendshipStatusNormal)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrFriendshipNotFound
	}
	return nil
}

func (r *FriendRepository) BlockUser(ctx context.Context, blockerID int64, blockedID int64) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO block_relations (blocker_id, blocked_id)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE blocker_id = blocker_id
`, blockerID, blockedID)
	return err
}

func (r *FriendRepository) UnblockUser(ctx context.Context, blockerID int64, blockedID int64) error {
	_, err := r.db.ExecContext(ctx, `
DELETE FROM block_relations
WHERE blocker_id = ? AND blocked_id = ?
`, blockerID, blockedID)
	return err
}

func (r *FriendRepository) IsBlocked(ctx context.Context, blockerID int64, blockedID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
SELECT 1
FROM block_relations
WHERE blocker_id = ? AND blocked_id = ?
LIMIT 1
`, blockerID, blockedID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanFriendRequestWithUser(row scanner) (*model.FriendRequestWithUser, error) {
	var item model.FriendRequestWithUser
	if err := row.Scan(
		&item.Request.ID,
		&item.Request.RequestID,
		&item.Request.FromUserID,
		&item.Request.ToUserID,
		&item.Request.Message,
		&item.Request.Status,
		&item.Request.CreatedAt,
		&item.Request.UpdatedAt,
		&item.User.User.ID,
		&item.User.User.UserID,
		&item.User.User.Username,
		&item.User.User.PasswordHash,
		&item.User.User.UserType,
		&item.User.User.Status,
		&item.User.User.CreatedAt,
		&item.User.User.UpdatedAt,
		&item.User.User.DeletedAt,
		&item.User.Profile.ID,
		&item.User.Profile.UserID,
		&item.User.Profile.Nickname,
		&item.User.Profile.AvatarURL,
		&item.User.Profile.Gender,
		&item.User.Profile.Bio,
		&item.User.Profile.ProfileStatus,
		&item.User.Profile.ProfileReviewReason,
		&item.User.Profile.CreatedAt,
		&item.User.Profile.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanFriendshipWithUser(row scanner) (*model.FriendshipWithUser, error) {
	var item model.FriendshipWithUser
	if err := row.Scan(
		&item.Friendship.ID,
		&item.Friendship.UserID1,
		&item.Friendship.UserID2,
		&item.Friendship.Status,
		&item.Friendship.CreatedAt,
		&item.Friendship.UpdatedAt,
		&item.User.User.ID,
		&item.User.User.UserID,
		&item.User.User.Username,
		&item.User.User.PasswordHash,
		&item.User.User.UserType,
		&item.User.User.Status,
		&item.User.User.CreatedAt,
		&item.User.User.UpdatedAt,
		&item.User.User.DeletedAt,
		&item.User.Profile.ID,
		&item.User.Profile.UserID,
		&item.User.Profile.Nickname,
		&item.User.Profile.AvatarURL,
		&item.User.Profile.Gender,
		&item.User.Profile.Bio,
		&item.User.Profile.ProfileStatus,
		&item.User.Profile.ProfileReviewReason,
		&item.User.Profile.CreatedAt,
		&item.User.Profile.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func orderedFriendPair(userID1 int64, userID2 int64) (int64, int64) {
	if userID1 < userID2 {
		return userID1, userID2
	}
	return userID2, userID1
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *drivermysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
