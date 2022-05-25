package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	errorsx "github.com/hamba/pkg/v2/errors"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/nrwiersma/aura/pkg/store"
)

// ErrNotFound is returned when a record is not found.
const ErrNotFound = errorsx.Error("not found")

// Registry represents an image registry.
type Registry interface {
	Resolve(ctx context.Context, img image.Image) (image.Image, error)
	ExtractProcfile(ctx context.Context, img string) ([]byte, error)
}

// Store represents a data store.
type Store interface {
	App(ctx context.Context, q store.AppsQuery) (*store.App, error)
	Apps(ctx context.Context, q store.AppsQuery) ([]*store.App, error)
	CreateApp(ctx context.Context, app *store.App) (*store.App, error)
	UpdateApp(ctx context.Context, app *store.App) error
	DeleteApp(ctx context.Context, app *store.App) error
	Release(ctx context.Context, q store.ReleasesQuery) (*store.Release, error)
	Releases(ctx context.Context, q store.ReleasesQuery) ([]*store.Release, error)
	CreateRelease(ctx context.Context, release *store.Release) (*store.Release, error)
}

// Controller manages the deployment of applications.
type Controller struct {
	reg   Registry
	store Store
}

// New returns a controller.
func New(reg Registry, store Store) *Controller {
	return &Controller{
		reg:   reg,
		store: store,
	}
}

// App contains the info of an application.
type App struct {
	ID        string
	Name      string
	CreatedAt *time.Time
	DeletedAt *time.Time
}

func toApp(app *store.App) *App {
	if app == nil {
		return nil
	}
	return &App{
		ID:        app.ID,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		DeletedAt: app.DeletedAt,
	}
}

func fromApp(app *App) *store.App {
	return &store.App{
		ID:        app.ID,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		DeletedAt: app.DeletedAt,
	}
}

// AppsQuery contains an applications query.
type AppsQuery struct {
	ID string

	Name string
}

// App returns the first application matching the app query.
func (a *Controller) App(ctx context.Context, q AppsQuery) (*App, error) {
	app, err := a.store.App(ctx, store.AppsQuery{ID: q.ID, Name: q.Name})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("could not find app: %w", err)
		}
	}
	return toApp(app), nil
}

// Apps returns all applications matching the app query.
func (a *Controller) Apps(ctx context.Context, q AppsQuery) ([]*App, error) {
	apps, err := a.store.Apps(ctx, store.AppsQuery{ID: q.ID, Name: q.Name})
	if err != nil {
		return nil, fmt.Errorf("could not get apps: %w", err)
	}

	res := make([]*App, 0, len(apps))
	for _, app := range apps {
		res = append(res, toApp(app))
	}
	return res, nil
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
func (a *Controller) Create(ctx context.Context, cfg CreateConfig) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, ValidationError{err: err}
	}

	app, err := a.store.CreateApp(ctx, &store.App{
		Name: cfg.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create app: %w", err)
	}
	return toApp(app), nil
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
func (a *Controller) Destroy(ctx context.Context, cfg DestroyConfig) error {
	if err := cfg.Validate(); err != nil {
		return ValidationError{err: err}
	}

	if err := a.store.DeleteApp(ctx, fromApp(cfg.App)); err != nil {
		return fmt.Errorf("could not delete app: %w", err)
	}

	// TODO: Remove the application from the deployment.

	return nil
}

// Release contains the info for a release.
type Release struct {
	ID        string
	App       *App
	Image     *image.Image
	Version   int
	Procfile  []byte
	CreatedAt *time.Time
}

func toRelease(release *store.Release) *Release {
	if release == nil {
		return nil
	}
	res := &Release{
		ID:        release.ID,
		App:       toApp(release.App),
		Image:     release.Image,
		Version:   release.Version,
		Procfile:  release.Procfile,
		CreatedAt: release.CreatedAt,
	}
	return res
}

// ReleasesQuery contains a release query.
type ReleasesQuery struct {
	App *App

	Version int
}

// Release returns the first application matching the query.
func (a *Controller) Release(ctx context.Context, q ReleasesQuery) (*Release, error) {
	release, err := a.store.Release(ctx, store.ReleasesQuery{App: fromApp(q.App), Version: q.Version})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return toRelease(release), nil
}

// Releases returns all releases matching the query.
func (a *Controller) Releases(ctx context.Context, q ReleasesQuery) ([]*Release, error) {
	release, err := a.store.Releases(ctx, store.ReleasesQuery{App: fromApp(q.App), Version: q.Version})
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	res := make([]*Release, 0, len(release))
	for _, release := range release {
		res = append(res, toRelease(release))
	}
	return res, nil
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
func (a *Controller) Deploy(ctx context.Context, cfg DeployConfig) (*Release, error) {
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

	release, err := a.store.CreateRelease(ctx, &store.Release{
		AppID:    cfg.App.ID,
		Image:    &img,
		Procfile: procFile,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create release: %w", err)
	}

	// TODO: Deploy the release with the deployment.

	return toRelease(release), nil
}

// ValidationError is returned when there is a validation error.
type ValidationError struct {
	err error
}

// Error stringifies the error.
func (e ValidationError) Error() string {
	if e.err == nil {
		return "validation error"
	}
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e ValidationError) Unwrap() error { return e.err }
