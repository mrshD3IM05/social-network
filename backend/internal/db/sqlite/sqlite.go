package sqlite

import (
	"database/sql"
	"errors"
	"strings"

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
	// foreign_keys and busy_timeout are per-connection settings. Running them
	// once with db.Exec only configures whichever pooled connection served that
	// call, so they go in the DSN instead: the driver then applies them to every
	// connection it opens. journal_mode is stored in the file itself.
	db, err := sql.Open("sqlite3", dsn(path))
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, err
	}

	if err := verifyForeignKeys(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func dsn(path string) string {
	if strings.Contains(path, "?") {
		return path + "&_foreign_keys=on&_busy_timeout=5000"
	}
	return path + "?_foreign_keys=on&_busy_timeout=5000"
}

// verifyForeignKeys fails fast if the DSN did not take effect, rather than
// letting the app run with silent referential-integrity gaps.
func verifyForeignKeys(db *sql.DB) error {
	var enabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		return err
	}
	if enabled != 1 {
		return errors.New("sqlite: foreign key enforcement is off")
	}
	return nil
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
