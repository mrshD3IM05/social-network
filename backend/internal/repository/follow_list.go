package repository

import (
	"time"

	"sn-backend/internal/model"
)

// PendingFollowRequest carries the requester's public identity alongside the
// request id, so a client can render an accept/decline row without a second call.
type PendingFollowRequest struct {
	ID        int64
	Status    string
	CreatedAt time.Time
	From      *model.User
}

const followUserColumns = `u.id, u.first_name, u.last_name, u.nickname, u.avatar, u.about_me, u.private, u.created_at`

func scanFollowUser(s scanner) (*model.User, error) {
	user := new(model.User)
	var private int
	if err := s.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Nickname, &user.Avatar, &user.AboutMe, &private, &user.CreatedAt); err != nil {
		return nil, err
	}
	user.Private = private == 1
	return user, nil
}

func (r *Repository) collectUsers(query string, args ...any) ([]*model.User, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*model.User, 0)
	for rows.Next() {
		user, err := scanFollowUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// ListFollowers returns the accepted followers of userID.
func (r *Repository) ListFollowers(userID int64) ([]*model.User, error) {
	return r.collectUsers(
		`SELECT `+followUserColumns+`
		 FROM follow_requests f JOIN users u ON u.id = f.from_user_id
		 WHERE f.to_user_id = ? AND f.status = ?
		 ORDER BY u.first_name, u.last_name, u.id`,
		userID, model.FollowAccepted,
	)
}

// ListFollowing returns the users userID follows with an accepted request.
func (r *Repository) ListFollowing(userID int64) ([]*model.User, error) {
	return r.collectUsers(
		`SELECT `+followUserColumns+`
		 FROM follow_requests f JOIN users u ON u.id = f.to_user_id
		 WHERE f.from_user_id = ? AND f.status = ?
		 ORDER BY u.first_name, u.last_name, u.id`,
		userID, model.FollowAccepted,
	)
}

// ListAllUsers lists every account except the caller, newest profiles last.
func (r *Repository) ListAllUsers(exceptID int64) ([]*model.User, error) {
	return r.collectUsers(
		`SELECT `+followUserColumns+` FROM users u WHERE u.id <> ? ORDER BY u.first_name, u.last_name, u.id`,
		exceptID,
	)
}

// ListPendingFollowRequests lists the still-undecided requests addressed to userID.
func (r *Repository) ListPendingFollowRequests(userID int64) ([]*PendingFollowRequest, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.status, f.created_at, `+followUserColumns+`
		 FROM follow_requests f JOIN users u ON u.id = f.from_user_id
		 WHERE f.to_user_id = ? AND f.status = ?
		 ORDER BY f.created_at DESC, f.id DESC`,
		userID, model.FollowPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]*PendingFollowRequest, 0)
	for rows.Next() {
		request := new(PendingFollowRequest)
		user := new(model.User)
		var private int
		if err := rows.Scan(
			&request.ID, &request.Status, &request.CreatedAt,
			&user.ID, &user.FirstName, &user.LastName, &user.Nickname, &user.Avatar, &user.AboutMe, &private, &user.CreatedAt,
		); err != nil {
			return nil, err
		}
		user.Private = private == 1
		request.From = user
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

// FollowState reports how viewerID relates to targetID: "", "pending",
// "accepted" or "declined".
func (r *Repository) FollowState(viewerID, targetID int64) (string, error) {
	var status string
	err := r.QueryRow(
		`SELECT status FROM follow_requests WHERE from_user_id = ? AND to_user_id = ?`,
		viewerID, targetID,
	).Scan(&status)
	if err != nil {
		if notFound(err) == ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return status, nil
}
