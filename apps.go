package aura

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

// App contains the info of an application.
type App struct {
	ID        string
	Name      string
	CreatedAt *time.Time
	DeletedAt *time.Time
}

// BeforeCreate is a pre-creation hook.
func (a *App) BeforeCreate(_ *gorm.DB) error {
	a.ID = ksuid.New().String()

	now := time.Now().UTC()
	a.CreatedAt = &now

	return nil
}

type appService struct {
	db *DB
}

func (s *appService) First(ctx context.Context, scope scope) (*App, error) {
	var app *App
	scope = composedScope{order("name"), scope}
	return app, s.db.WithContext(ctx).Scopes(scope.scope).First(&app).Error
}

func (s *appService) Find(ctx context.Context, scope scope) ([]*App, error) {
	var apps []*App
	scope = composedScope{order("name"), scope}
	return apps, s.db.WithContext(ctx).Scopes(scope.scope).Find(&apps).Error
}

func (s *appService) Create(ctx context.Context, app *App) (*App, error) {
	return app, s.db.WithContext(ctx).Create(app).Error
}

func (s *appService) Update(ctx context.Context, app *App) error {
	return s.db.WithContext(ctx).Save(app).Error
}

func (s *appService) Delete(ctx context.Context, app *App) error {
	now := time.Now().UTC()
	app.DeletedAt = &now
	return s.Update(ctx, app)
}
