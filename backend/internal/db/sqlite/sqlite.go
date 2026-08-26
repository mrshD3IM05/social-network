package sqlite

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	migrateiofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"sn-backend/internal/db/migrations"
)

var DB *sql.DB

func InitDB(path string) error {
	db, err := Open(path)
	if err != nil {
		return err
	}

	if err := Migrate(db); err != nil {
		return err
	}

	DB = db
	return nil
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	for _, stmt := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	source, err := migrateiofs.New(migrations.FS, "sqlite")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	source.Close()
	return nil
}
