package docker_test

import (
	"archive/tar"
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	httptest "github.com/hamba/testutils/http"
	"github.com/nrwiersma/aura/pkg/docker"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Resolve(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/images/create").ReturnsString(http.StatusOK, `{}`)
	srv.On(http.MethodGet, "/images/some/repo:latest/json").ReturnsString(http.StatusOK, `{"RepoDigests":["some/repo@sha256:c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2"]}}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	got, err := reg.Resolve(context.Background(), image.Image{
		Repository: "some/repo",
		Tag:        "latest",
	})

	want := image.Image{
		Repository: "some/repo",
		Digest:     "sha256:c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
	}
	require.NoError(t, err)
	assert.Equal(t, want, got)
	srv.AssertExpectations()
}

func TestRegistry_ResolveHandlesDigest(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/images/create").ReturnsString(http.StatusOK, `{}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	got, err := reg.Resolve(context.Background(), image.Image{
		Repository: "some/repo",
		Digest:     "sha256:c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
	})

	want := image.Image{
		Repository: "some/repo",
		Digest:     "sha256:c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
	}
	require.NoError(t, err)
	assert.Equal(t, want, got)
	srv.AssertExpectations()
}

func TestRegistry_ResolveHandlesNoDigest(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/images/create").ReturnsString(http.StatusOK, `{}`)
	srv.On(http.MethodGet, "/images/some/repo:latest/json").ReturnsString(http.StatusOK, `{"RepoDigests":[]}}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	got, err := reg.Resolve(context.Background(), image.Image{
		Repository: "some/repo",
		Tag:        "latest",
	})

	want := image.Image{
		Repository: "some/repo",
		Tag:        "latest",
	}
	require.NoError(t, err)
	assert.Equal(t, want, got)
	srv.AssertExpectations()
}

func TestRegistry_ResolveHandlesPullError(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/images/create").Returns(http.StatusInternalServerError, nil)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	_, err = reg.Resolve(context.Background(), image.Image{
		Repository: "some/repo",
		Tag:        "latest",
	})

	require.Error(t, err)
	srv.AssertExpectations()
}

func TestRegistry_ResolveHandlesInspectError(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/images/create").ReturnsString(http.StatusOK, `{}`)
	srv.On(http.MethodGet, "/images/some/repo:latest/json").Returns(http.StatusInternalServerError, nil)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	_, err = reg.Resolve(context.Background(), image.Image{
		Repository: "some/repo",
		Tag:        "latest",
	})

	require.Error(t, err)
	srv.AssertExpectations()
}

func TestRegistry_ExtractProcfile(t *testing.T) {
	procFile := "web: test"

	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/containers/create").ReturnsString(http.StatusOK, `{"ID": "foo"}`)
	srv.On(http.MethodGet, "/containers/foo/json").ReturnsString(http.StatusOK, `{}`)
	srv.On(http.MethodGet, "/containers/foo/archive").Handle(func(rw http.ResponseWriter, req *http.Request) {
		path := req.URL.Query().Get("path")
		assert.Equal(t, "Procfile", path)

		_, _ = rw.Write(tarFile(t, "Procfile", procFile))
	})
	srv.On(http.MethodDelete, "/containers/foo").ReturnsString(http.StatusOK, `{}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	got, err := reg.ExtractProcfile(context.Background(), "foo/bar:latest")

	require.NoError(t, err)
	assert.Equal(t, []byte(procFile), got)
	srv.AssertExpectations()
}

func TestRegistry_ExtractProcfileHandlesWorkingDirectory(t *testing.T) {
	procFile := "web: test"

	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/containers/create").ReturnsString(http.StatusOK, `{"ID": "foo"}`)
	srv.On(http.MethodGet, "/containers/foo/json").ReturnsString(http.StatusOK, `{"Config":{"WorkingDir":"/app"}}`)
	srv.On(http.MethodGet, "/containers/foo/archive").Handle(func(rw http.ResponseWriter, req *http.Request) {
		path := req.URL.Query().Get("path")
		assert.Equal(t, "/app/Procfile", path)

		_, _ = rw.Write(tarFile(t, "Procfile", procFile))
	})
	srv.On(http.MethodDelete, "/containers/foo").ReturnsString(http.StatusOK, `{}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	got, err := reg.ExtractProcfile(context.Background(), "foo/bar:latest")

	require.NoError(t, err)
	assert.Equal(t, []byte(procFile), got)
	srv.AssertExpectations()
}

func TestRegistry_ExtractProcfileHandlesCreateContainerError(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/containers/create").ReturnsString(http.StatusInternalServerError, ``)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	_, err = reg.ExtractProcfile(context.Background(), "foo/bar:latest")

	require.Error(t, err)
	srv.AssertExpectations()
}

func TestRegistry_ExtractProcfileHandlesInspectError(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/containers/create").ReturnsString(http.StatusOK, `{"ID": "foo"}`)
	srv.On(http.MethodGet, "/containers/foo/json").ReturnsString(http.StatusInternalServerError, ``)
	srv.On(http.MethodDelete, "/containers/foo").ReturnsString(http.StatusOK, `{}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	_, err = reg.ExtractProcfile(context.Background(), "nrwiersma/test:latest")

	require.Error(t, err)
	srv.AssertExpectations()
}

func TestRegistry_ExtractProcfileHandlesDownloadError(t *testing.T) {
	srv := httptest.NewServer(t)
	srv.On(http.MethodPost, "/containers/create").ReturnsString(http.StatusOK, `{"ID": "foo"}`)
	srv.On(http.MethodGet, "/containers/foo/json").ReturnsString(http.StatusOK, `{}`)
	srv.On(http.MethodGet, "/containers/foo/archive").Returns(http.StatusInternalServerError, nil)
	srv.On(http.MethodDelete, "/containers/foo").ReturnsString(http.StatusOK, `{}`)

	cancel := swapEnv(t, "DOCKER_HOST", srv.URL())
	t.Cleanup(cancel)

	reg, err := docker.NewRegistry()
	require.NoError(t, err)

	_, err = reg.ExtractProcfile(context.Background(), "foo/bar:latest")

	require.Error(t, err)
	srv.AssertExpectations()
}

func swapEnv(t *testing.T, key, value string) (cancel func()) {
	t.Helper()

	oldVal := os.Getenv(key)

	err := os.Setenv(key, value)
	require.NoError(t, err)

	return func() {
		_ = os.Setenv(key, oldVal)
	}
}

func tarFile(t *testing.T, name, value string) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: name,
		Size: int64(len(value)),
	}
	err := tw.WriteHeader(hdr)
	require.NoError(t, err)
	_, err = tw.Write([]byte(value))
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}
