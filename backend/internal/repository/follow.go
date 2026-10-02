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
// every call, and start after the user lastID (0: the first page). The viewer
// is who is asking, which is not necessarily the subject: both rows carry the
// relation the viewer has with them.
func (r *Repository) ListFollowers(viewerID, userID, lastID int64) ([]*model.User, error) {
	return r.listFollowUsers(viewerID,
		`SELECT `+userColumns+viewerStateColumns+`
		 FROM follow_requests f
		 JOIN `+userTable+` ON v.id = f.from_user_id
		 `+viewerStateJoins+`
		 WHERE f.to_user_id = ? AND f.status = ? AND `+afterUserCondition+`
		 ORDER BY v.first_name COLLATE NOCASE, v.last_name COLLATE NOCASE, v.id
		 LIMIT ?`,
		userID, model.FollowAccepted, lastID, lastID, PageSize,
	)
}

func (r *Repository) ListFollowing(viewerID, userID, lastID int64) ([]*model.User, error) {
	return r.listFollowUsers(viewerID,
		`SELECT `+userColumns+viewerStateColumns+`
		 FROM follow_requests f
		 JOIN `+userTable+` ON v.id = f.to_user_id
		 `+viewerStateJoins+`
		 WHERE f.from_user_id = ? AND f.status = ? AND `+afterUserCondition+`
		 ORDER BY v.first_name COLLATE NOCASE, v.last_name COLLATE NOCASE, v.id
		 LIMIT ?`,
		userID, model.FollowAccepted, lastID, lastID, PageSize,
	)
}

// ListMessageableUsers returns the users userID can start a private chat with.
// The rule is the one CanMessage applies before every message: at least one of
// the two follows the other, with an accepted request — in either direction, so
// a pending follow counts for nothing. Yourself is left out, like in /users.
// Ordered by name so the Messages list reads like the people directory. The
// caller is the viewer, so the relation on each row is the one they have with
// the person.
func (r *Repository) ListMessageableUsers(userID int64) ([]*model.User, error) {
	return r.listFollowUsers(userID,
		`SELECT `+userColumns+viewerStateColumns+`
		 FROM `+userTable+`
		 `+viewerStateJoins+`
		 WHERE v.id != ?
		   AND EXISTS (
				SELECT 1
				FROM follow_requests f
				WHERE (
					(f.from_user_id = ? AND f.to_user_id = v.id)
					OR
					(f.from_user_id = v.id AND f.to_user_id = ?)
				)
				AND f.status = ?
		   )
		 ORDER BY (
				SELECT MAX(m.created_at)
				FROM messages m
				WHERE
					(m.from_user_id = ? AND m.to_user_id = v.id)
					OR
					(m.from_user_id = v.id AND m.to_user_id = ?)
		   ) DESC,
		   v.first_name COLLATE NOCASE,
		   v.last_name COLLATE NOCASE,
		   v.id`,
		userID,
		userID, userID, model.FollowAccepted,
		userID, userID,
	)
}

// listFollowUsers runs one of the people queries above. The query has to be
// written with viewerStateJoins and viewerStateColumns, and the viewer goes in
// ahead of every argument it carries, because those two placeholders are the
// first ones in the statement.
func (r *Repository) listFollowUsers(viewerID int64, query string, args ...any) ([]*model.User, error) {
	rows, err := r.db.Query(query, append([]any{viewerID, viewerID}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		user, err := scanUserForViewer(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// ListPendingFollowRequests returns the follow requests waiting for userID to
// accept or decline them, with the user who sent each one, carrying the
// relation userID has with that sender.
func (r *Repository) ListPendingFollowRequests(viewerID, userID int64) ([]*model.FollowRequest, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.created_at, `+userColumns+viewerStateColumns+`
		 FROM follow_requests f
		 JOIN `+userTable+` ON v.id = f.from_user_id
		 `+viewerStateJoins+`
		 WHERE f.to_user_id = ? AND f.status = ?
		 ORDER BY f.id DESC`,
		viewerID, viewerID, userID, model.FollowPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*model.FollowRequest{}
	for rows.Next() {
		request := &model.FollowRequest{ToUserID: userID, Status: model.FollowPending}
		// the request id and time come first, then the sender as the caller
		// relates to them
		from, err := scanUserRow(rows, []any{&request.ID, &request.CreatedAt}, true)
		if err != nil {
			return nil, err
		}
		request.From = from
		request.FromUserID = from.ID
		requests = append(requests, request)
	}
	return requests, rows.Err()
}
