package repository

import (
	"errors"

	"sn-backend/internal/model"
)

func (r *Repository) CreateGroup(group *model.Group) error {
	if group == nil {
		return errors.New("group is nil")
	}

	result, err := r.db.Exec(
		`INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)`,
		group.OwnerID,
		group.Name,
		group.Description,
	)
	if err != nil {
		return err
	}

	group.ID, err = result.LastInsertId()
	return err
}

func (r *Repository) GetGroup(id int64) (*model.Group, error) {
	group := new(model.Group)
	err := r.QueryRow(
		`SELECT id, creator_id, title, description, created_at FROM groups WHERE id = ?`,
		id,
	).Scan(&group.ID, &group.OwnerID, &group.Name, &group.Description, &group.CreatedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return group, nil
}

func (r *Repository) UpdateGroup(group *model.Group) error {
	if group == nil {
		return errors.New("group is nil")
	}

	_, err := r.db.Exec(
		`UPDATE groups SET creator_id = ?, title = ?, description = ? WHERE id = ?`,
		group.OwnerID,
		group.Name,
		group.Description,
		group.ID,
	)
	return err
}

func (r *Repository) DeleteGroup(id int64) error {
	_, err := r.db.Exec("DELETE FROM groups WHERE id = ?", id)
	return err
}
