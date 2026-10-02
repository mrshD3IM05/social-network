package repository

import "sn-backend/internal/model"

func (r *Repository) GetFollowRequest(fromUserID, toUserID int64) (*model.FollowRequest, error) {
	follow := new(model.FollowRequest)
	err := r.QueryRow(
		`SELECT id, from_user_id, to_user_id, status, created_at
		 FROM follow_requests WHERE from_user_id = ? AND to_user_id = ?`,
		fromUserID, toUserID,
	).Scan(&follow.ID, &follow.FromUserID, &follow.ToUserID, &follow.Status, &follow.CreatedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return follow, nil
}

func (r *Repository) CreateFollowRequest(fromUserID, toUserID int64, status string) (*model.FollowRequest, error) {
	result, err := r.db.Exec(
		`INSERT INTO follow_requests (from_user_id, to_user_id, status) VALUES (?, ?, ?)`,
		fromUserID, toUserID, status,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestByID(id)
}

func (r *Repository) GetFollowRequestByID(id int64) (*model.FollowRequest, error) {
	follow := new(model.FollowRequest)
	err := r.QueryRow(
		`SELECT id, from_user_id, to_user_id, status, created_at FROM follow_requests WHERE id = ?`, id,
	).Scan(&follow.ID, &follow.FromUserID, &follow.ToUserID, &follow.Status, &follow.CreatedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return follow, nil
}

func (r *Repository) UpdateFollowStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE follow_requests SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *Repository) DeleteFollow(fromUserID, toUserID int64) error {
	_, err := r.db.Exec(
		`DELETE FROM follow_requests WHERE from_user_id = ? AND to_user_id = ?`,
		fromUserID, toUserID,
	)
	return err
}

func (r *Repository) IsFollowing(fromUserID, toUserID int64) (bool, error) {
	var exists int
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM follow_requests WHERE from_user_id = ? AND to_user_id = ? AND status = ?)`,
		fromUserID, toUserID, model.FollowAccepted,
	).Scan(&exists)
	return exists == 1, err
}

// ListFollowers returns the users who follow userID, ListFollowing the users
// userID follows. Only accepted requests count — a pending one is not a follow.
// Both order by name like the people directory, so the lists read the same on
// every call, and start after the user lastID (0: the first page).
func (r *Repository) ListFollowers(userID, lastID int64) ([]*model.User, error) {
	return r.listFollowUsers(
		`SELECT `+userColumnsPrefixed+`
		 FROM follow_requests f
		 JOIN users u ON u.id = f.from_user_id
		 WHERE f.to_user_id = ? AND f.status = ? AND `+afterUserCondition+`
		 ORDER BY u.first_name COLLATE NOCASE, u.last_name COLLATE NOCASE, u.id
		 LIMIT ?`,
		userID, lastID,
	)
}

func (r *Repository) ListFollowing(userID, lastID int64) ([]*model.User, error) {
	return r.listFollowUsers(
		`SELECT `+userColumnsPrefixed+`
		 FROM follow_requests f
		 JOIN users u ON u.id = f.to_user_id
		 WHERE f.from_user_id = ? AND f.status = ? AND `+afterUserCondition+`
		 ORDER BY u.first_name COLLATE NOCASE, u.last_name COLLATE NOCASE, u.id
		 LIMIT ?`,
		userID, lastID,
	)
}

// ListMessageableUsers returns the users userID can start a private chat with.
// The rule is the one CanMessage applies before every message: at least one of
// the two follows the other, with an accepted request — in either direction, so
// a pending follow counts for nothing. Yourself is left out, like in /users.
// Ordered by name so the Messages list reads like the people directory.
func (r *Repository) ListMessageableUsers(userID int64) ([]*model.User, error) {
	return r.listFollowUsers(
		`SELECT `+userColumnsPrefixed+`
		 FROM users u
		 WHERE u.id != ?
		   AND EXISTS (
				SELECT 1
				FROM follow_requests f
				WHERE (
					(f.from_user_id = ? AND f.to_user_id = u.id)
					OR
					(f.from_user_id = u.id AND f.to_user_id = ?)
				)
				AND f.status = ?
		   )
		 ORDER BY (
				SELECT MAX(m.created_at)
				FROM messages m
				WHERE
					(m.from_user_id = ? AND m.to_user_id = u.id)
					OR
					(m.from_user_id = u.id AND m.to_user_id = ?)
		   ) DESC,
		   u.first_name COLLATE NOCASE,
		   u.last_name COLLATE NOCASE,
		   u.id`,
		userID,
		userID, userID, model.FollowAccepted,
		userID, userID,
	)
}

func (r *Repository) listFollowUsers(query string, args ...any) ([]*model.User, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// CountFollowers is how many users follow userID, CountFollowing how many
// users userID follows.
func (r *Repository) CountFollowers(userID int64) (int, error) {
	var count int
	err := r.QueryRow(`SELECT COUNT(*) FROM follow_requests WHERE to_user_id = ? AND status = ?`, userID, model.FollowAccepted).Scan(&count)
	return count, err
}

func (r *Repository) CountFollowing(userID int64) (int, error) {
	var count int
	err := r.QueryRow(`SELECT COUNT(*) FROM follow_requests WHERE from_user_id = ? AND status = ?`, userID, model.FollowAccepted).Scan(&count)
	return count, err
}

// ListPendingFollowRequests returns the follow requests waiting for userID to
// accept or decline them, with the user who sent each one.
func (r *Repository) ListPendingFollowRequests(userID int64) ([]*model.FollowRequest, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.created_at, `+userColumnsPrefixed+`
		 FROM follow_requests f
		 JOIN users u ON u.id = f.from_user_id
		 WHERE f.to_user_id = ? AND f.status = ?
		 ORDER BY f.id DESC`,
		userID, model.FollowPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*model.FollowRequest{}
	for rows.Next() {
		request := &model.FollowRequest{ToUserID: userID, Status: model.FollowPending, From: new(model.User)}
		var private int
		from := request.From
		if err := rows.Scan(&request.ID, &request.CreatedAt, &from.ID, &from.Email, &from.Password, &from.FirstName,
			&from.LastName, &from.DateOfBirth, &from.Avatar, &from.Nickname, &from.AboutMe, &private, &from.CreatedAt); err != nil {
			return nil, err
		}
		from.Private = private == 1
		request.FromUserID = from.ID
		requests = append(requests, request)
	}
	return requests, rows.Err()
}
