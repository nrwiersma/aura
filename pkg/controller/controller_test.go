package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nrwiersma/aura/pkg/controller"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/nrwiersma/aura/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestController_App(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name    string
		qry     controller.AppsQuery
		app     *store.App
		err     error
		want    *controller.App
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles getting an app",
			qry:     controller.AppsQuery{ID: "123"},
			app:     &store.App{ID: "123", Name: "test app", CreatedAt: &now},
			want:    &controller.App{ID: "123", Name: "test app", CreatedAt: &now},
			wantErr: require.NoError,
		},
		{
			name:    "handles app not found",
			qry:     controller.AppsQuery{ID: "123"},
			err:     store.ErrNotFound,
			wantErr: require.Error,
		},
		{
			name:    "handles store error",
			qry:     controller.AppsQuery{ID: "123"},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			db.On("App", store.AppsQuery{ID: test.qry.ID, Name: test.qry.Name}).Return(test.app, test.err)

			contr := controller.New(reg, db)

			got, err := contr.App(context.Background(), test.qry)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Apps(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name    string
		qry     controller.AppsQuery
		apps    []*store.App
		err     error
		want    []*controller.App
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles getting apps",
			qry:     controller.AppsQuery{ID: "123"},
			apps:    []*store.App{{ID: "123", Name: "test app", CreatedAt: &now}},
			want:    []*controller.App{{ID: "123", Name: "test app", CreatedAt: &now}},
			wantErr: require.NoError,
		},
		{
			name:    "handles store error",
			qry:     controller.AppsQuery{ID: "123"},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			db.On("Apps", store.AppsQuery{ID: test.qry.ID, Name: test.qry.Name}).Return(test.apps, test.err)

			contr := controller.New(reg, db)

			got, err := contr.Apps(context.Background(), test.qry)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Create(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name    string
		cfg     controller.CreateConfig
		app     *store.App
		err     error
		want    *controller.App
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles creating an app",
			cfg:     controller.CreateConfig{Name: "test app"},
			app:     &store.App{ID: "123", Name: "test app", CreatedAt: &now},
			want:    &controller.App{ID: "123", Name: "test app", CreatedAt: &now},
			wantErr: require.NoError,
		},
		{
			name:    "handles no name",
			cfg:     controller.CreateConfig{Name: ""},
			wantErr: require.Error,
		},
		{
			name:    "handles store error",
			cfg:     controller.CreateConfig{Name: "test app"},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			if test.app != nil || test.err != nil {
				db.On("CreateApp", &store.App{Name: test.cfg.Name}).Return(test.app, test.err)
			}

			contr := controller.New(reg, db)

			got, err := contr.Create(context.Background(), test.cfg)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Destroy(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name    string
		cfg     controller.DestroyConfig
		err     error
		wantApp *store.App
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles destroying an app",
			cfg:     controller.DestroyConfig{App: &controller.App{ID: "123", Name: "test app", CreatedAt: &now}},
			wantApp: &store.App{ID: "123", Name: "test app", CreatedAt: &now},
			wantErr: require.NoError,
		},
		{
			name:    "handles no app",
			cfg:     controller.DestroyConfig{},
			wantErr: require.Error,
		},
		{
			name:    "handles invalid app",
			cfg:     controller.DestroyConfig{App: &controller.App{}},
			wantErr: require.Error,
		},
		{
			name:    "handles store error",
			cfg:     controller.DestroyConfig{App: &controller.App{ID: "123", Name: "test app", CreatedAt: &now}},
			wantApp: &store.App{ID: "123", Name: "test app", CreatedAt: &now},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			if test.wantApp != nil {
				db.On("DeleteApp", test.wantApp).Return(test.err)
			}

			contr := controller.New(reg, db)

			err := contr.Destroy(context.Background(), test.cfg)

			test.wantErr(t, err)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Release(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name    string
		qry     controller.ReleasesQuery
		release *store.Release
		err     error
		want    *controller.Release
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles getting a release",
			qry:     controller.ReleasesQuery{App: &controller.App{ID: "123"}, Version: 1},
			release: &store.Release{ID: "123", App: &store.App{ID: "123"}, Version: 1, CreatedAt: &now},
			want:    &controller.Release{ID: "123", App: &controller.App{ID: "123"}, Version: 1, CreatedAt: &now},
			wantErr: require.NoError,
		},
		{
			name:    "handles release not found",
			qry:     controller.ReleasesQuery{App: &controller.App{ID: "123"}, Version: 1},
			err:     store.ErrNotFound,
			wantErr: require.Error,
		},
		{
			name:    "handles store error",
			qry:     controller.ReleasesQuery{App: &controller.App{ID: "123"}, Version: 1},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			db.On("Release", store.ReleasesQuery{App: fromApp(test.qry.App), Version: test.qry.Version}).Return(test.release, test.err)

			contr := controller.New(reg, db)

			got, err := contr.Release(context.Background(), test.qry)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Releases(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name     string
		qry      controller.ReleasesQuery
		releases []*store.Release
		err      error
		want     []*controller.Release
		wantErr  require.ErrorAssertionFunc
	}{
		{
			name:     "handles getting releases",
			qry:      controller.ReleasesQuery{App: &controller.App{ID: "123"}},
			releases: []*store.Release{{ID: "123", App: &store.App{ID: "123"}, Version: 1, CreatedAt: &now}},
			want:     []*controller.Release{{ID: "123", App: &controller.App{ID: "123"}, Version: 1, CreatedAt: &now}},
			wantErr:  require.NoError,
		},
		{
			name:    "handles store error",
			qry:     controller.ReleasesQuery{App: &controller.App{ID: "123"}},
			err:     errors.New("test"),
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			db := &mockStore{}
			db.On("Releases", store.ReleasesQuery{App: fromApp(test.qry.App), Version: test.qry.Version}).Return(test.releases, test.err)

			contr := controller.New(reg, db)

			got, err := contr.Releases(context.Background(), test.qry)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestController_Deploy(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name       string
		cfg        controller.DeployConfig
		resolveErr error
		extractErr error
		release    *store.Release
		releaseErr error
		want       *controller.Release
		wantErr    require.ErrorAssertionFunc
	}{
		{
			name:    "handles deploying a release",
			cfg:     controller.DeployConfig{App: &controller.App{ID: "123"}, Image: image.Image{Repository: "foo/bar"}},
			release: &store.Release{ID: "123", App: &store.App{ID: "123"}, Version: 1, CreatedAt: &now},
			want:    &controller.Release{ID: "123", App: &controller.App{ID: "123"}, Version: 1, CreatedAt: &now},
			wantErr: require.NoError,
		},
		{
			name:    "handles no app",
			cfg:     controller.DeployConfig{},
			wantErr: require.Error,
		},
		{
			name:    "handles invalid app",
			cfg:     controller.DeployConfig{App: &controller.App{}},
			wantErr: require.Error,
		},
		{
			name:       "handles resolve error",
			cfg:        controller.DeployConfig{App: &controller.App{ID: "123"}, Image: image.Image{Repository: "foo/bar"}},
			resolveErr: errors.New("test"),
			wantErr:    require.Error,
		},
		{
			name:       "handles extract error",
			cfg:        controller.DeployConfig{App: &controller.App{ID: "123"}, Image: image.Image{Repository: "foo/bar"}},
			extractErr: errors.New("test"),
			wantErr:    require.Error,
		},
		{
			name:       "handles store error",
			cfg:        controller.DeployConfig{App: &controller.App{ID: "123"}, Image: image.Image{Repository: "foo/bar"}},
			releaseErr: errors.New("test"),
			wantErr:    require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reg := &mockRegistry{}
			if test.cfg.App != nil && test.cfg.App.ID != "" {
				reg.On("Resolve", test.cfg.Image).Return(test.cfg.Image, test.resolveErr)
				if test.resolveErr == nil {
					reg.On("ExtractProcfile", test.cfg.Image.String()).Return([]byte("test"), test.extractErr)
				}
			}
			db := &mockStore{}
			if test.release != nil || test.releaseErr != nil {
				db.On("CreateRelease", &store.Release{AppID: test.cfg.App.ID, Image: &test.cfg.Image, Procfile: []byte("test")}).Return(test.release, test.releaseErr)
			}

			contr := controller.New(reg, db)

			got, err := contr.Deploy(context.Background(), test.cfg)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
			reg.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func fromApp(app *controller.App) *store.App {
	return &store.App{
		ID:        app.ID,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		DeletedAt: app.DeletedAt,
	}
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

type mockStore struct {
	mock.Mock
}

func (m *mockStore) App(_ context.Context, q store.AppsQuery) (*store.App, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.App), args.Error(1)
}

func (m *mockStore) Apps(_ context.Context, q store.AppsQuery) ([]*store.App, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*store.App), args.Error(1)
}

func (m *mockStore) CreateApp(_ context.Context, app *store.App) (*store.App, error) {
	args := m.Called(app)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.App), args.Error(1)
}

func (m *mockStore) UpdateApp(_ context.Context, app *store.App) error {
	args := m.Called(app)
	return args.Error(0)
}

func (m *mockStore) DeleteApp(_ context.Context, app *store.App) error {
	args := m.Called(app)
	return args.Error(0)
}

func (m *mockStore) Release(_ context.Context, q store.ReleasesQuery) (*store.Release, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Release), args.Error(1)
}

func (m *mockStore) Releases(_ context.Context, q store.ReleasesQuery) ([]*store.Release, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*store.Release), args.Error(1)
}

func (m *mockStore) CreateRelease(_ context.Context, release *store.Release) (*store.Release, error) {
	args := m.Called(release)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Release), args.Error(1)
}
