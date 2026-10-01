package repository

import (
	"errors"
	"strings"

	"sn-backend/internal/model"
)

const userColumns = "id, email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, private, created_at"

// Same columns, qualified with the users alias, for queries that join another table.
const userColumnsPrefixed = "u.id, u.email, u.password, u.first_name, u.last_name, u.date_of_birth, u.avatar, u.nickname, u.about_me, u.private, u.created_at"

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

// afterUserCondition keeps the users u that come after the user lastID in
// name order (first name, last name, id), so a page starts right after the
// one before. It takes lastID twice: 0 means the first page.
const afterUserCondition = `(? = 0 OR (u.first_name COLLATE NOCASE, u.last_name COLLATE NOCASE, u.id) >
	(SELECT first_name, last_name, id FROM users WHERE id = ?))`

// ListUsers returns one page of the users except excludeID whose name or
// nickname contains search (empty search keeps everyone), ordered by name,
// starting after the user lastID.
func (r *Repository) ListUsers(excludeID int64, search string, lastID int64) ([]*model.User, error) {
	// % and _ are wildcards in LIKE, so they are escaped to be searched as text
	pattern := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search) + "%"
	rows, err := r.db.Query(
		`SELECT `+userColumnsPrefixed+` FROM users u
		 WHERE u.id != ? AND (u.first_name || ' ' || u.last_name || ' ' || u.nickname) LIKE ? ESCAPE '\'
		 AND `+afterUserCondition+`
		 ORDER BY u.first_name COLLATE NOCASE, u.last_name COLLATE NOCASE, u.id
		 LIMIT ?`,
		excludeID, pattern, lastID, lastID, PageSize,
	)
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

// SetUserPrivate flips only the privacy column, so turning a profile
// public or private cannot touch any other field.
func (r *Repository) SetUserPrivate(id int64, private bool) error {
	_, err := r.db.Exec(`UPDATE users SET private = ? WHERE id = ?`, boolToInt(private), id)
	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
