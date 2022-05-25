package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nrwiersma/aura/pkg/image"
	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReleasesQuery contains a release query.
type ReleasesQuery struct {
	App *App

	Version int
}

func (q ReleasesQuery) scope(db *gorm.DB) *gorm.DB {
	var scope composedScope

	if q.App != nil {
		scope = append(scope, fieldEquals("app_id", q.App.ID))
	}

	if q.Version > 0 {
		scope = append(scope, fieldEquals("version", q.Version))
	}

	return scope.scope(db)
}

// Release contains the info for a release.
type Release struct {
	ID        string
	AppID     string
	App       *App
	Image     *image.Image
	Version   int
	Procfile  []byte
	CreatedAt *time.Time
}

// BeforeCreate is a pre-creation hook.
func (r *Release) BeforeCreate(_ *gorm.DB) error {
	r.ID = ksuid.New().String()

	now := time.Now().UTC()
	r.CreatedAt = &now

	return nil
}

var releasesPreload = preload("App")

// Release returns the first release for the given query.
func (db *Database) Release(ctx context.Context, q ReleasesQuery) (*Release, error) {
	var release *Release
	scope := composedScope{releasesPreload, order("version"), q}
	if err := db.db.WithContext(ctx).Scopes(scope.scope).First(&release).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return release, nil
}

// Releases returns the releases for the given query.
func (db *Database) Releases(ctx context.Context, q ReleasesQuery) ([]*Release, error) {
	var releases []*Release
	scope := composedScope{releasesPreload, order("version"), q}
	return releases, db.db.WithContext(ctx).Scopes(scope.scope).Find(&releases).Error
}

// CreateRelease creates a release.
func (db *Database) CreateRelease(ctx context.Context, release *Release) (*Release, error) {
	tx := db.db.WithContext(ctx).Begin()
	defer func() { _ = tx.Rollback() }()

	// Lock the releases for this app for updates, so the version number if unique in a race.
	_ = tx.Where("app_id = ?", release.AppID).Clauses(clause.Locking{Strength: "UPDATE"}).Error

	ver, err := db.currentVersion(tx, release.AppID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			// It is ok if it doesn't exist, we will use version 1 then.
		default:
			return nil, fmt.Errorf("getting latest release version: %w", err)
		}
	}
	release.Version = ver + 1

	if err = tx.Create(release).Error; err != nil {
		return nil, fmt.Errorf("creating release: %w", err)
	}

	return release, tx.Commit().Error
}

func (db *Database) currentVersion(tx *gorm.DB, appID string) (int, error) {
	var release *Release
	return release.Version, tx.Where("app_id = ?", appID).Order("version DESC").First(&release).Error
}
