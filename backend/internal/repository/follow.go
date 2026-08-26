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
