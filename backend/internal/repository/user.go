package repository

import (
	"errors"

	"sn-backend/internal/model"
)

const userColumns = "id, email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, private, created_at"

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (*model.User, error) {
	user := new(model.User)
	var private int

	if err := s.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&private,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	user.Private = private == 1
	return user, nil
}

func (r *Repository) CreateUser(user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	result, err := r.db.Exec(
		`INSERT INTO users (email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, private) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.Avatar,
		user.Nickname,
		user.AboutMe,
		boolToInt(user.Private),
	)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

func (r *Repository) GetUserByID(id int64) (*model.User, error) {
	user, err := scanUser(r.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
	if err != nil {
		return nil, notFound(err)
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(email string) (*model.User, error) {
	user, err := scanUser(r.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = ?`, email))
	if err != nil {
		return nil, notFound(err)
	}
	return user, nil
}

func (r *Repository) GetUserByNickname(nickname string) (*model.User, error) {
	user, err := scanUser(r.QueryRow(`SELECT `+userColumns+` FROM users WHERE nickname = ?`, nickname))
	if err != nil {
		return nil, notFound(err)
	}
	return user, nil
}

func (r *Repository) UpdateUser(user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	_, err := r.db.Exec(
		`UPDATE users SET first_name = ?, last_name = ?, date_of_birth = ?, avatar = ?, nickname = ?, about_me = ?, private = ? WHERE id = ?`,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.Avatar,
		user.Nickname,
		user.AboutMe,
		boolToInt(user.Private),
		user.ID,
	)
	return err
}

func (r *Repository) DeleteUser(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
