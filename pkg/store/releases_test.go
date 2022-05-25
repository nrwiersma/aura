package store_test

import (
	"context"
	"testing"

	"github.com/nrwiersma/aura/pkg/image"
	"github.com/nrwiersma/aura/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAura_CreateRelease(t *testing.T) {
	db := testDB(t)

	img, err := image.Decode("foo/bar:latest")
	require.NoError(t, err)

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})

	got, err := db.CreateRelease(context.Background(), &store.Release{App: app, Image: &img, Procfile: []byte("test")})

	require.NoError(t, err)
	assert.Equal(t, "foo/bar:latest", got.Image.String())
	assert.Equal(t, 1, got.Version)
	assert.Equal(t, []byte("test"), got.Procfile)
	assert.NotNil(t, app.ID)
	assert.NotNil(t, app.CreatedAt)
}

func TestAura_Releases(t *testing.T) {
	db := testDB(t)

	img := image.Image{Repository: "foo/bar", Tag: "latest"}

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})
	require.NoError(t, err)

	release1, err := db.CreateRelease(context.Background(), &store.Release{AppID: app.ID, Image: &img, Procfile: []byte("test")})
	require.NoError(t, err)
	release2, err := db.CreateRelease(context.Background(), &store.Release{AppID: app.ID, Image: &img, Procfile: []byte("test")})
	require.NoError(t, err)

	release1.App = app
	release2.App = app

	got, err := db.Releases(context.Background(), store.ReleasesQuery{})

	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, []*store.Release{release1, release2}, got)
}

func TestAura_Release(t *testing.T) {
	db := testDB(t)

	img := image.Image{Repository: "foo/bar", Tag: "latest"}

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})
	require.NoError(t, err)

	want, err := db.CreateRelease(context.Background(), &store.Release{AppID: app.ID, Image: &img, Procfile: []byte("test")})
	require.NoError(t, err)
	want.App = app

	got, err := db.Release(context.Background(), store.ReleasesQuery{App: app, Version: 1})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestAura_ReleaseHandlesNotFound(t *testing.T) {
	db := testDB(t)

	app, err := db.CreateApp(context.Background(), &store.App{Name: "test app"})
	require.NoError(t, err)

	_, err = db.Release(context.Background(), store.ReleasesQuery{App: app, Version: 1})

	require.Error(t, err)
}
