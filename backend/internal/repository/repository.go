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

// PageSize is how many items a list endpoint answers at once ("10 by 10").
// The next ones are asked for with ?last=<id of the last item already shown>.
const PageSize = 10

// MessagePageSize is the chat history page: 10 like every other list, older
// ones asked for with ?last=<id of the oldest message shown>.
const MessagePageSize = PageSize

// dbtx is what repository queries run against: the shared pool, or one
// transaction bound to a call to Repositories.WithinTx.
type dbtx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type dbStore struct {
	db dbtx
}

type Repositories struct {
	store         *dbStore
	Users         *UserRepository
	Follows       *FollowRepository
	Posts         *PostRepository
	Reactions     *ReactionRepository
	Comments      *CommentRepository
	Events        *EventRepository
	Files         *FileRepository
	Groups        *GroupRepository
	Messages      *MessageRepository
	Notifications *NotificationRepository
	Sessions      *SessionRepository
}

type UserRepository struct{ *dbStore }
type FollowRepository struct{ *dbStore }
type PostRepository struct {
	*dbStore
	comments  *CommentRepository
	reactions *ReactionRepository
}
type ReactionRepository struct{ *dbStore }
type CommentRepository struct{ *dbStore }
type EventRepository struct{ *dbStore }
type FileRepository struct{ *dbStore }
type GroupRepository struct {
	*dbStore
	users *UserRepository
}
type MessageRepository struct{ *dbStore }
type NotificationRepository struct{ *dbStore }
type SessionRepository struct{ *dbStore }

func New(db *sql.DB) *Repositories {
	return newRepositories(&dbStore{db: db})
}

// newRepositories binds a fresh set of repositories to one store, so the same
// wiring serves the pool and a transaction.
func newRepositories(store *dbStore) *Repositories {
	users := &UserRepository{dbStore: store}
	reactions := &ReactionRepository{dbStore: store}
	comments := &CommentRepository{dbStore: store}
	return &Repositories{
		store:         store,
		Users:         users,
		Follows:       &FollowRepository{dbStore: store},
		Posts:         &PostRepository{dbStore: store, comments: comments, reactions: reactions},
		Reactions:     reactions,
		Comments:      comments,
		Events:        &EventRepository{dbStore: store},
		Files:         &FileRepository{dbStore: store},
		Groups:        &GroupRepository{dbStore: store, users: users},
		Messages:      &MessageRepository{dbStore: store},
		Notifications: &NotificationRepository{dbStore: store},
		Sessions:      &SessionRepository{dbStore: store},
	}
}

// WithinTx runs fn against repositories bound to a single transaction, so a
// service can write across more than one repository atomically. Any error rolls
// every write back.
func (r *Repositories) WithinTx(fn func(tx *Repositories) error) error {
	return r.store.transaction(func(tx *dbStore) error {
		return fn(newRepositories(tx))
	})
}

func (r *dbStore) QueryRow(query string, args ...any) *sql.Row {
	return r.db.QueryRow(query, args...)
}

// begin starts a transaction on the underlying pool. A store already bound to a
// transaction cannot nest one.
func (r *dbStore) begin() (*sql.Tx, error) {
	db, ok := r.db.(*sql.DB)
	if !ok {
		return nil, errors.New("repository: transaction already active")
	}
	return db.Begin()
}

// transaction runs fn inside a transaction. When the store is already bound to a
// transaction (e.g. inside Repositories.WithinTx) it joins that transaction
// instead of opening a new one, so repository operations can be composed.
func (r *dbStore) transaction(fn func(tx *dbStore) error) error {
	if tx, ok := r.db.(*sql.Tx); ok {
		return fn(&dbStore{db: tx})
	}
	tx, err := r.begin()
	if err != nil {
		return err
	}
	if err := fn(&dbStore{db: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
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
