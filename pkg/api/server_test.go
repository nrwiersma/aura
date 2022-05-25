package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hamba/logger/v2"
	"github.com/hamba/statter/v2"
	"github.com/nrwiersma/aura/pkg/api"
	"github.com/nrwiersma/aura/pkg/controller"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func requireDoRequest(t *testing.T, method, url string, body []byte) *http.Response {
	t.Helper()

	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, r)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func requireJSONEncode(t *testing.T, obj interface{}) []byte {
	t.Helper()

	b, err := json.Marshal(obj)
	require.NoError(t, err)

	return b
}

func requireJSONDecode(t *testing.T, r io.Reader, obj interface{}) {
	t.Helper()

	err := json.NewDecoder(r).Decode(obj)
	require.NoError(t, err)
}

func setupTestServer(t *testing.T, app api.Delegate) string {
	t.Helper()

	log := logger.New(io.Discard, logger.LogfmtFormat(), logger.Error)
	stats := statter.New(statter.DiscardReporter, time.Minute)

	apiSrv := api.New(app, log, stats)

	server := httptest.NewServer(apiSrv)
	t.Cleanup(server.Close)

	return server.URL
}

type mockApp struct {
	mock.Mock
}

func (m *mockApp) App(_ context.Context, q controller.AppsQuery) (*controller.App, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller.App), args.Error(1)
}

func (m *mockApp) Apps(_ context.Context, q controller.AppsQuery) ([]*controller.App, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*controller.App), args.Error(1)
}

func (m *mockApp) Create(_ context.Context, cfg controller.CreateConfig) (*controller.App, error) {
	args := m.Called(cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller.App), args.Error(1)
}

func (m *mockApp) Destroy(_ context.Context, cfg controller.DestroyConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *mockApp) Release(_ context.Context, q controller.ReleasesQuery) (*controller.Release, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller.Release), args.Error(1)
}

func (m *mockApp) Releases(_ context.Context, q controller.ReleasesQuery) ([]*controller.Release, error) {
	args := m.Called(q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*controller.Release), args.Error(1)
}

func (m *mockApp) Deploy(_ context.Context, cfg controller.DeployConfig) (*controller.Release, error) {
	args := m.Called(cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller.Release), args.Error(1)
}
