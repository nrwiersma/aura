package store

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

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

// App gets the first app that matches the query.
func (db *Database) App(ctx context.Context, q AppsQuery) (*App, error) {
	var app *App
	scope := composedScope{order("name"), q}
	if err := db.db.WithContext(ctx).Scopes(scope.scope).First(&app).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return app, nil
}

// Apps gets all apps matching the query.
func (db *Database) Apps(ctx context.Context, q AppsQuery) ([]*App, error) {
	var apps []*App
	scope := composedScope{order("name"), q}
	return apps, db.db.WithContext(ctx).Scopes(scope.scope).Find(&apps).Error
}

// CreateApp creates an application.
func (db *Database) CreateApp(ctx context.Context, app *App) (*App, error) {
	return app, db.db.WithContext(ctx).Create(app).Error
}

// UpdateApp updates an application.
func (db *Database) UpdateApp(ctx context.Context, app *App) error {
	return db.db.WithContext(ctx).Save(app).Error
}

// DeleteApp deletes an app.
func (db *Database) DeleteApp(ctx context.Context, app *App) error {
	now := time.Now().UTC()
	app.DeletedAt = &now
	return db.UpdateApp(ctx, app)
}
