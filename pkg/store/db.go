package store

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/hamba/logger/v2"
	errorsx "github.com/hamba/pkg/v2/errors"
	"github.com/nrwiersma/aura/pkg/migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// ErrNotFound is returned when a record is not found.
const ErrNotFound = errorsx.Error("not found")

// Database is a database store.
type Database struct {
	db *gorm.DB

	schema   Schema
	migrator *migrate.Migrator
}

// OpenDB returns a connected database.
func OpenDB(dsn string, log *logger.Logger) (*Database, error) {
	if _, err := url.Parse(dsn); err != nil {
		return nil, fmt.Errorf("could not parse db dsn: %w", err)
	}

	return NewDatabase(postgres.Open(dsn), log)
}

// NewDatabase return a DB from the given connection.
func NewDatabase(dialect gorm.Dialector, log *logger.Logger) (*Database, error) {
	db, err := gorm.Open(dialect, &gorm.Config{
		Logger: glogger.New(logAdapter{log: log}, glogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  glogger.Warn,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not resolve to db: %w", err)
	}

	return &Database{
		db:       db,
		migrator: migrate.New(sqlDB, new(sync.Mutex)),
	}, nil
}

// Migrate migrates the database up.
func (db *Database) Migrate() error {
	err := db.migrator.Run(db.schema.migrations()...)
	if err != nil {
		return fmt.Errorf("could not migrate the database: %w", err)
	}
	return nil
}

// IsHealthy determines if the database is healthy.
func (db *Database) IsHealthy() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

type logAdapter struct {
	log *logger.Logger
}

func (l logAdapter) Printf(s string, args ...interface{}) {
	l.log.Debug(fmt.Sprintf(s, args...))
}

type scope interface {
	scope(*gorm.DB) *gorm.DB
}

type scopeFunc func(*gorm.DB) *gorm.DB

func (fn scopeFunc) scope(db *gorm.DB) *gorm.DB {
	return fn(db)
}

type composedScope []scope

func (s composedScope) scope(db *gorm.DB) *gorm.DB {
	for _, scope := range s {
		scope.scope(db)
	}
	return db
}

func idEquals(id string) scope {
	return fieldEquals("id", id)
}

func fieldEquals(field string, v interface{}) scope {
	return scopeFunc(func(db *gorm.DB) *gorm.DB {
		return db.Where(field+" = ?", v)
	})
}

func isNull(field string) scope {
	return scopeFunc(func(db *gorm.DB) *gorm.DB {
		return db.Where(field + " is null")
	})
}

func preload(associations ...string) scope {
	scope := make(composedScope, 0, len(associations))
	for _, a := range associations {
		aa := a
		scope = append(scope, scopeFunc(func(db *gorm.DB) *gorm.DB {
			return db.Preload(aa)
		}))
	}

	return scope
}

func order(field string) scope {
	return scopeFunc(func(db *gorm.DB) *gorm.DB {
		return db.Order(field)
	})
}
