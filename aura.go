package aura

import (
	"context"
	"errors"
	"fmt"

	errorsx "github.com/hamba/pkg/v2/errors"
	"github.com/nrwiersma/aura/pkg/image"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a record is not found.
const ErrNotFound = errorsx.Error("not found")

// Registry represents an image registry.
type Registry interface {
	Resolve(ctx context.Context, img image.Image) (image.Image, error)
	ExtractProcfile(ctx context.Context, img string) ([]byte, error)
}

// Aura manages the deployment of applications.
type Aura struct {
	db  *DB
	reg Registry

	apps     *appService
	releases *releaseService
}

// New returns an app handler.
func New(db *DB, reg Registry) *Aura {
	aura := &Aura{
		db:  db,
		reg: reg,
	}

	aura.apps = &appService{db: db}
	aura.releases = &releaseService{db: db}

	return aura
}

// AppsQuery contains an applications query.
type AppsQuery struct {
	ID string

	Name string
}

func (q AppsQuery) scope(db *gorm.DB) *gorm.DB {
	scope := composedScope{isNull("deleted_at")}

	if q.ID != "" {
		scope = append(scope, idEquals(q.ID))
	}

	if q.Name != "" {
		scope = append(scope, fieldEquals("name", q.Name))
	}

	return scope.scope(db)
}

// AppsFind returns the first application matching the app query.
func (a *Aura) AppsFind(ctx context.Context, q AppsQuery) (*App, error) {
	app, err := a.apps.First(ctx, q)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("could not find app: %w", err)
		}
	}
	return app, nil
}

// Apps returns all applications matching the app query.
func (a *Aura) Apps(ctx context.Context, q AppsQuery) ([]*App, error) {
	return a.apps.Find(ctx, q)
}

// CreateConfig contains application creation configuration.
type CreateConfig struct {
	Name string
}

// Validate validates a create configuration.
func (c CreateConfig) Validate() error {
	if c.Name == "" {
		return errors.New("app name is required")
	}

	return nil
}

// Create creates an application.
func (a *Aura) Create(ctx context.Context, cfg CreateConfig) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, ValidationError{err: err}
	}

	app, err := a.apps.Create(ctx, &App{
		Name: cfg.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create app: %w", err)
	}
	return app, nil
}

// DestroyConfig contains application removal configuration.
type DestroyConfig struct {
	App *App
}

// Validate validates a destroy configuration.
func (c DestroyConfig) Validate() error {
	if c.App == nil {
		return errors.New("an application is required")
	}
	if c.App.ID == "" {
		return errors.New("the application is invalid")
	}

	return nil
}

// Destroy removes an application and any existing deployment.
func (a *Aura) Destroy(ctx context.Context, cfg DestroyConfig) error {
	if err := cfg.Validate(); err != nil {
		return ValidationError{err: err}
	}

	if err := a.apps.Delete(ctx, cfg.App); err != nil {
		return fmt.Errorf("could not delete app: %w", err)
	}

	// TODO: Remove the application from the deployment.

	return nil
}

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

// ReleasesFind returns the first application matching the query.
func (a *Aura) ReleasesFind(ctx context.Context, q ReleasesQuery) (*Release, error) {
	release, err := a.releases.First(ctx, q)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return release, nil
}

// Releases returns all releases matching the query.
func (a *Aura) Releases(ctx context.Context, q ReleasesQuery) ([]*Release, error) {
	return a.releases.Find(ctx, q)
}

// DeployConfig contains application release configuration.
type DeployConfig struct {
	App *App

	Image image.Image
}

// Validate validates a deploy configuration.
func (c DeployConfig) Validate() error {
	if c.App == nil {
		return errors.New("an application is required")
	}
	if c.App.ID == "" {
		return errors.New("the application is invalid")
	}

	return nil
}

// Deploy creates a release and deploys it.
func (a *Aura) Deploy(ctx context.Context, cfg DeployConfig) (*Release, error) {
	if err := cfg.Validate(); err != nil {
		return nil, ValidationError{err: err}
	}

	img, err := a.reg.Resolve(ctx, cfg.Image)
	if err != nil {
		return nil, fmt.Errorf("could not resolve image: %w", err)
	}

	procFile, err := a.reg.ExtractProcfile(ctx, img.String())
	if err != nil {
		return nil, fmt.Errorf("could not extract procfile: %w", err)
	}

	release, err := a.releases.Create(ctx, &Release{
		AppID:    cfg.App.ID,
		Image:    &img,
		Procfile: procFile,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create release: %w", err)
	}

	// TODO: Deploy the release with the deployment.

	return release, nil
}
