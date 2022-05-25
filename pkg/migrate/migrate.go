package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
)

const table = "schema_migrations"

// MigrationError is returned when a migration has an error.
type MigrationError struct {
	err error
}

// Error stringifies the error.
func (e MigrationError) Error() string {
	if e.err == nil {
		return "migration error"
	}
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e MigrationError) Unwrap() error { return e.err }

// Migration contains a migration.
type Migration struct {
	ID   int
	Up   func(tx *sql.Tx) error
	Down func(tx *sql.Tx) error
}

// Migrator runs database migrations.
type Migrator struct {
	db     *sql.DB
	locker sync.Locker
}

// New returns a migrator.
func New(db *sql.DB, locker sync.Locker) *Migrator {
	return &Migrator{
		db:     db,
		locker: locker,
	}
}

// Run runs the migrations.
func (m *Migrator) Run(migrations ...Migration) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	_, err := m.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (version integer PRIMARY KEY NOT NULL)`, table))
	if err != nil {
		return fmt.Errorf("could not create migrations table: %w", err)
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	sortMigrations(migrations)
	for _, migration := range migrations {
		if err = m.runMigration(tx, migration); err != nil {
			return fmt.Errorf("could not run migration %d: %w", migration.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit migration transaction: %w", err)
	}
	return nil
}

func (m *Migrator) runMigration(tx *sql.Tx, migration Migration) error {
	ok, err := m.shouldMigrate(tx, migration.ID)
	if err != nil {
		return fmt.Errorf("determining if migration has been run: %w", err)
	}
	if !ok {
		return nil
	}

	if err = migration.Up(tx); err != nil {
		return MigrationError{err: err}
	}

	_, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (version) VALUES (%d)", table, migration.ID))
	if err != nil {
		return fmt.Errorf("inserting migration marker: %w", err)
	}
	return nil
}

func (m *Migrator) shouldMigrate(tx *sql.Tx, id int) (bool, error) {
	var i int
	err := tx.QueryRow(fmt.Sprintf("SELECT 1 FROM %s WHERE version = %d", table, id)).Scan(&i)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return errors.Is(err, sql.ErrNoRows), nil
}

func sortMigrations(m []Migration) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].ID < m[j].ID
	})
}

// Queries creates a migration function from a set of queries.
func Queries(qrys ...string) func(*sql.Tx) error {
	return func(tx *sql.Tx) error {
		for _, qry := range qrys {
			if _, err := tx.Exec(qry); err != nil {
				return err
			}
		}
		return nil
	}
}
