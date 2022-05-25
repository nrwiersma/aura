package store_test

import (
	"context"
	"testing"

	"github.com/nrwiersma/aura/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateApp(t *testing.T) {
	db := testDB(t)

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})

	require.NoError(t, err)
	assert.Equal(t, "test app", app.Name)
	assert.NotNil(t, app.ID)
	assert.NotNil(t, app.CreatedAt)
}

func TestDatabase_Apps(t *testing.T) {
	db := testDB(t)

	app1, err := db.CreateApp(context.Background(), &store.App{Name: "test1 app"})
	require.NoError(t, err)
	app2, err := db.CreateApp(context.Background(), &store.App{Name: "test2 app"})
	require.NoError(t, err)

	apps, err := db.Apps(context.Background(), store.AppsQuery{})

	require.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, []*store.App{app1, app2}, apps)
}

func TestDatabase_App(t *testing.T) {
	db := testDB(t)

	want, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})
	require.NoError(t, err)

	got, err := db.App(context.Background(), store.AppsQuery{ID: want.ID})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDatabase_AppByName(t *testing.T) {
	db := testDB(t)

	want, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})
	require.NoError(t, err)

	got, err := db.App(context.Background(), store.AppsQuery{Name: "test app"})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDatabase_AppHandlesNotFound(t *testing.T) {
	db := testDB(t)

	_, err := db.App(context.Background(), store.AppsQuery{Name: "test app"})

	require.Error(t, err)
}

func TestDatabase_DeleteApp(t *testing.T) {
	db := testDB(t)

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})

	err = db.DeleteApp(context.Background(), app)

	require.NoError(t, err)
}
