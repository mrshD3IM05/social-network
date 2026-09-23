package repository

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNotFound = errors.New("repository: not found")
	ErrExists   = errors.New("repository: already exists")
	ErrNotOwner = errors.New("repository: not the owner")
)

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

// placeholders returns "?, ?, ?" for n arguments (IN clauses).
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// int64sToAny converts an ID slice into query arguments.
func int64sToAny(ids []int64) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
