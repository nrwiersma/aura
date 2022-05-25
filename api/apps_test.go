package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/nrwiersma/aura"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_HandleGetApps(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		apps           []*aura.App
		err            error
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			apps:           []*aura.App{{ID: "test", Name: "test app", CreatedAt: &now}},
			wantStatusCode: http.StatusOK,
			wantResp:       `[{"id":"test","name":"test app","createdAt":"2022-02-01T04:00:00Z"}]`,
		},
		{
			name:           "handles app error",
			err:            errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			app := &mockApp{}
			app.On("Apps", aura.AppsQuery{}).Return(test.apps, test.err)

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodGet, srvUrl+"/apps", nil)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}

func TestServer_HandleGetApp(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		app            *aura.App
		err            error
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			app:            &aura.App{ID: "test", Name: "test app", CreatedAt: &now},
			wantStatusCode: http.StatusOK,
			wantResp:       `{"id":"test","name":"test app","createdAt":"2022-02-01T04:00:00Z"}`,
		},
		{
			name:           "handles app not found",
			err:            aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"app not found"}`,
		},
		{
			name:           "handles app error",
			err:            errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			app := &mockApp{}
			app.On("AppsFind", aura.AppsQuery{ID: "123"}).Return(test.app, test.err)

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodGet, srvUrl+"/apps/123", nil)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}

func TestServer_HandleCreateApp(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	type req struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name           string
		req            string
		app            *aura.App
		err            error
		wantAppName    string
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			req:            `{"name":"test app"}`,
			app:            &aura.App{ID: "123", Name: "test app", CreatedAt: &now},
			wantAppName:    "test app",
			wantStatusCode: http.StatusOK,
			wantResp:       `{"id":"123","name":"test app","createdAt":"2022-02-01T04:00:00Z"}`,
		},
		{
			name:           "handles invalid json",
			req:            `{"name":"test app}`,
			wantStatusCode: http.StatusBadRequest,
			wantResp:       `{"error":"invalid app data"}`,
		},
		{
			name:           "handles validation error",
			req:            `{"name":"test app"}`,
			err:            aura.ValidationError{},
			wantAppName:    "test app",
			wantStatusCode: http.StatusBadRequest,
			wantResp:       `{"error":"invalid app: validation error"}`,
		},
		{
			name:           "handles app error",
			req:            `{"name":"test app"}`,
			err:            errors.New("test"),
			wantAppName:    "test app",
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			app := &mockApp{}
			if test.wantAppName != "" {
				app.On("Create", aura.CreateConfig{Name: test.wantAppName}).Return(test.app, test.err)
			}

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodPost, srvUrl+"/apps/", []byte(test.req))
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}

func TestServer_HandleDestroyApp(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		app            *aura.App
		findErr        error
		destroyErr     error
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			app:            &aura.App{ID: "test", Name: "test app", CreatedAt: &now},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "handles app not found",
			findErr:        aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"app not found"}`,
		},
		{
			name:           "handles app error",
			app:            &aura.App{ID: "test", Name: "test app", CreatedAt: &now},
			destroyErr:     errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			app := &mockApp{}
			app.On("AppsFind", aura.AppsQuery{ID: "123"}).Return(test.app, test.findErr)
			if test.app != nil {
				app.On("Destroy", aura.DestroyConfig{App: test.app}).Return(test.destroyErr)
			}

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodDelete, srvUrl+"/apps/123", nil)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}
