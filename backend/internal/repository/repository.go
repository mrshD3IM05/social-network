package repository

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("repository: not found")

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) QueryRow(query string, args ...any) *sql.Row {
	return r.db.QueryRow(query, args...)
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
