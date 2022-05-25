package aura

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

type releaseService struct {
	db *DB
}

func (s *releaseService) First(ctx context.Context, scope scope) (*Release, error) {
	var release *Release
	scope = composedScope{releasesPreload, order("version"), scope}
	return release, s.db.WithContext(ctx).Scopes(scope.scope).First(&release).Error
}

func (s *releaseService) Find(ctx context.Context, scope scope) ([]*Release, error) {
	var releases []*Release
	scope = composedScope{releasesPreload, order("version"), scope}
	return releases, s.db.WithContext(ctx).Scopes(scope.scope).Find(&releases).Error
}

func (s *releaseService) Create(ctx context.Context, release *Release) (*Release, error) {
	tx := s.db.WithContext(ctx).Begin()
	defer func() { _ = tx.Rollback() }()

	// Lock the releases for this app for updates, so the version number if unique in a race.
	_ = tx.Where("app_id = ?", release.AppID).Clauses(clause.Locking{Strength: "UPDATE"}).Error

	ver, err := s.currentVersion(tx, release.AppID)
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

func (s *releaseService) currentVersion(tx *gorm.DB, appID string) (int, error) {
	var release *Release
	return release.Version, tx.Where("app_id = ?", appID).Order("version DESC").First(&release).Error
}
