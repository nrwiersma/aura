package aura_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/hamba/logger/v2"
	"github.com/nrwiersma/aura"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
)

func TestAura_Create(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		wantName string
		wantErr  require.ErrorAssertionFunc
	}{
		{
			name:     "handles creating an app",
			appName:  "test app",
			wantName: "test app",
			wantErr:  require.NoError,
		},
		{
			name:    "handles no name",
			appName: "",
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			db := testDB(t)
			reg := &mockRegistry{}

			a := aura.New(db, reg)

			app, err := a.Create(context.Background(), aura.CreateConfig{Name: test.appName})

			test.wantErr(t, err)
			if test.wantName != "" {
				assert.Equal(t, test.wantName, app.Name)
				assert.NotNil(t, app.ID)
				assert.NotNil(t, app.CreatedAt)
			}
		})
	}
}

func TestAura_Apps(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	app1, err := a.Create(context.Background(), aura.CreateConfig{Name: "test1 app"})
	require.NoError(t, err)
	app2, err := a.Create(context.Background(), aura.CreateConfig{Name: "test2 app"})
	require.NoError(t, err)

	apps, err := a.Apps(context.Background(), aura.AppsQuery{})

	require.NoError(t, err)
	assert.Len(t, apps, 2)
	assert.Equal(t, []*aura.App{app1, app2}, apps)
}

func TestAura_AppsFind(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	want, err := a.Create(context.Background(), aura.CreateConfig{Name: "test app"})
	require.NoError(t, err)

	got, err := a.AppsFind(context.Background(), aura.AppsQuery{ID: want.ID})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestAura_AppsFindByName(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	want, err := a.Create(context.Background(), aura.CreateConfig{Name: "test app"})
	require.NoError(t, err)

	got, err := a.AppsFind(context.Background(), aura.AppsQuery{Name: "test app"})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestAura_AppsHandlesNoApp(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	_, err := a.AppsFind(context.Background(), aura.AppsQuery{Name: "test app"})

	require.Error(t, err)
}

func TestAura_Destroy(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	app, err := a.Create(context.Background(), aura.CreateConfig{Name: "test11 app"})
	require.NoError(t, err)

	err = a.Destroy(context.Background(), aura.DestroyConfig{App: app})

	require.NoError(t, err)
	got, err := a.Apps(context.Background(), aura.AppsQuery{ID: app.ID})
	assert.Len(t, got, 0)
}

func TestAura_DestroyHandlesBadConfig(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	err := a.Destroy(context.Background(), aura.DestroyConfig{})

	require.Error(t, err)
}

func TestAura_DestroyHandlesInvalidConfigApp(t *testing.T) {
	db := testDB(t)
	reg := &mockRegistry{}

	a := aura.New(db, reg)

	err := a.Destroy(context.Background(), aura.DestroyConfig{App: &aura.App{}})

	require.Error(t, err)
}

func TestAura_Deploy(t *testing.T) {
	tests := []struct {
		name         string
		image        string
		resolveErr   error
		procfile     []byte
		extractErr   error
		wantImage    string
		wantVersion  int
		wantProcfile []byte
		wantErr      require.ErrorAssertionFunc
	}{
		{
			name:         "handles creating an app",
			image:        "foo/bar:latest",
			procfile:     []byte("test"),
			wantImage:    "foo/bar:latest",
			wantVersion:  1,
			wantProcfile: []byte("test"),
			wantErr:      require.NoError,
		},
		{
			name:       "handles resolve error",
			image:      "foo/bar:latest",
			resolveErr: errors.New("test"),
			wantErr:    require.Error,
		},
		{
			name:       "handles extracting procfile error",
			image:      "foo/bar:latest",
			procfile:   []byte("test"),
			extractErr: errors.New("test"),
			wantErr:    require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			img, err := image.Decode(test.image)
			require.NoError(t, err)

			db := testDB(t)
			reg := &mockRegistry{}
			reg.On("Resolve", img).Return(img, test.resolveErr)
			if test.procfile != nil {
				reg.On("ExtractProcfile", img.String()).Return(test.procfile, test.extractErr)
			}

			a := aura.New(db, reg)

			app, err := a.Create(context.Background(), aura.CreateConfig{Name: "test app"})

			got, err := a.Deploy(context.Background(), aura.DeployConfig{App: app, Image: img})

			test.wantErr(t, err)
			if test.wantVersion != 0 {
				assert.Equal(t, test.wantImage, got.Image.String())
				assert.Equal(t, test.wantVersion, got.Version)
				assert.Equal(t, test.wantProcfile, got.Procfile)
				assert.NotNil(t, app.ID)
				assert.NotNil(t, app.CreatedAt)
			}
		})
	}
}

func TestAura_DeployHandlesValidationError(t *testing.T) {
	tests := []struct {
		name string
		app  *aura.App
	}{
		{
			name: "handles no app",
			app:  nil,
		},
		{
			name: "handles invalid app",
			app:  &aura.App{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			img, err := image.Decode("foo/bar:latest")
			require.NoError(t, err)

			db := testDB(t)
			reg := &mockRegistry{}

			a := aura.New(db, reg)

			_, err = a.Deploy(context.Background(), aura.DeployConfig{App: test.app, Image: img})

			assert.Error(t, err)
		})
	}
}

func TestAura_Releases(t *testing.T) {
	img := image.Image{Repository: "foo/bar", Tag: "latest"}

	db := testDB(t)
	reg := &mockRegistry{}
	reg.On("Resolve", img).Return(img, nil)
	reg.On("ExtractProcfile", "foo/bar:latest").Return([]byte("test"), nil)

	a := aura.New(db, reg)

	app, err := a.Create(context.Background(), aura.CreateConfig{Name: "test app"})
	require.NoError(t, err)

	release1, err := a.Deploy(context.Background(), aura.DeployConfig{App: app, Image: img})
	require.NoError(t, err)
	release2, err := a.Deploy(context.Background(), aura.DeployConfig{App: app, Image: img})
	require.NoError(t, err)

	release1.App = app
	release2.App = app

	got, err := a.Releases(context.Background(), aura.ReleasesQuery{})

	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, []*aura.Release{release1, release2}, got)
}

func TestAura_ReleasesFind(t *testing.T) {
	img := image.Image{Repository: "foo/bar", Tag: "latest"}

	db := testDB(t)
	reg := &mockRegistry{}
	reg.On("Resolve", img).Return(img, nil)
	reg.On("ExtractProcfile", "foo/bar:latest").Return([]byte("test"), nil)

	a := aura.New(db, reg)

	app, err := a.Create(context.Background(), aura.CreateConfig{Name: "test app"})
	require.NoError(t, err)
	want, err := a.Deploy(context.Background(), aura.DeployConfig{App: app, Image: img})
	require.NoError(t, err)
	want.App = app

	got, err := a.ReleasesFind(context.Background(), aura.ReleasesQuery{App: app, Version: 1})

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func testDB(t *testing.T) *aura.DB {
	t.Helper()

	log := logger.New(io.Discard, logger.LogfmtFormat(), logger.Debug)

	db, err := aura.NewDB(sqlite.Open("file::memory:"), log)
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}

type mockRegistry struct {
	mock.Mock
}

func (m *mockRegistry) Resolve(_ context.Context, img image.Image) (image.Image, error) {
	args := m.Called(img)
	return args.Get(0).(image.Image), args.Error(1)
}

func (m *mockRegistry) ExtractProcfile(_ context.Context, img string) ([]byte, error) {
	args := m.Called(img)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
