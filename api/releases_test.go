package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/nrwiersma/aura"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_HandleGetReleases(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		appErr         error
		releases       []*aura.Release
		releasesErr    error
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			releases:       []*aura.Release{{ID: "test", AppID: "123", Version: 2}},
			wantStatusCode: http.StatusOK,
			wantResp:       `[{"id":"test","app":{"id":"","name":"","createdAt":null},"image":"","version":2,"procfile":"","createdAt":null}]`,
		},
		{
			name:           "handles app not found",
			appErr:         aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"app not found"}`,
		},
		{
			name:           "handles app find error",
			appErr:         errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
		{
			name:           "handles releases error",
			releasesErr:    errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			a := &aura.App{ID: "123", Name: "test app", CreatedAt: &now}

			app := &mockApp{}
			app.On("AppsFind", aura.AppsQuery{ID: "123"}).Return(a, test.appErr)
			if test.releases != nil || test.releasesErr != nil {
				app.On("Releases", aura.ReleasesQuery{App: a}).Return(test.releases, test.releasesErr)
			}

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodGet, srvUrl+"/apps/123/releases", nil)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}

func TestServer_HandleGetRelease(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		appErr         error
		release        *aura.Release
		releaseErr     error
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			release:        &aura.Release{ID: "test", AppID: "123", Version: 2},
			wantStatusCode: http.StatusOK,
			wantResp:       `{"id":"test","app":{"id":"","name":"","createdAt":null},"image":"","version":2,"procfile":"","createdAt":null}`,
		},
		{
			name:           "handles app not found",
			appErr:         aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"app not found"}`,
		},
		{
			name:           "handles app find error",
			appErr:         errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
		{
			name:           "handles release not found error",
			releaseErr:     aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"release not found"}`,
		},
		{
			name:           "handles release error",
			releaseErr:     errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			a := &aura.App{ID: "123", Name: "test app", CreatedAt: &now}

			app := &mockApp{}
			app.On("AppsFind", aura.AppsQuery{ID: "123"}).Return(a, test.appErr)
			if test.release != nil || test.releaseErr != nil {
				app.On("ReleasesFind", aura.ReleasesQuery{App: a, Version: 2}).Return(test.release, test.releaseErr)
			}

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodGet, srvUrl+"/apps/123/releases/2", nil)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}

func TestServer_HandleDeployApp(t *testing.T) {
	now := time.Date(2022, 02, 01, 04, 00, 0, 0, time.UTC)

	tests := []struct {
		name           string
		req            string
		appErr         error
		release        *aura.Release
		releaseErr     error
		wantImage      string
		wantStatusCode int
		wantResp       string
	}{
		{
			name:           "handles request",
			req:            `{"image":"foo/bar:latest"}`,
			release:        &aura.Release{ID: "test", AppID: "123", Version: 2},
			wantImage:      "foo/bar:latest",
			wantStatusCode: http.StatusOK,
			wantResp:       `{"id":"test","app":{"id":"","name":"","createdAt":null},"image":"","version":2,"procfile":"","createdAt":null}`,
		},
		{
			name:           "handles invalid json",
			req:            `{"image":"foo/bar:latest}`,
			wantStatusCode: http.StatusBadRequest,
			wantResp:       `{"error":"invalid app deployment data"}`,
		},
		{
			name:           "handles empty image",
			req:            `{"image":""}`,
			wantStatusCode: http.StatusBadRequest,
			wantResp:       `{"error":"invalid app deploy: invalid image format"}`,
		},
		{
			name:           "handles app not found",
			req:            `{"image":"foo/bar:latest"}`,
			appErr:         aura.ErrNotFound,
			wantStatusCode: http.StatusNotFound,
			wantResp:       `{"error":"app not found"}`,
		},
		{
			name:           "handles app find error",
			req:            `{"image":"foo/bar:latest"}`,
			appErr:         errors.New("test"),
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
		{
			name:           "handles release error",
			req:            `{"image":"foo/bar:latest"}`,
			releaseErr:     errors.New("test"),
			wantImage:      "foo/bar:latest",
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			a := &aura.App{ID: "123", Name: "test app", CreatedAt: &now}

			app := &mockApp{}
			app.On("AppsFind", aura.AppsQuery{ID: "123"}).Maybe().Return(a, test.appErr)

			if test.wantImage != "" {
				img, err := image.Decode(test.wantImage)
				require.NoError(t, err)

				app.On("Deploy", aura.DeployConfig{App: a, Image: img}).Return(test.release, test.releaseErr)
			}

			srvUrl := setupTestServer(t, app)

			resp := requireDoRequest(t, http.MethodPost, srvUrl+"/apps/123/deploys", []byte(test.req))
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.Equal(t, test.wantStatusCode, resp.StatusCode)

			got, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.wantResp, string(got))
			app.AssertExpectations(t)
		})
	}
}
