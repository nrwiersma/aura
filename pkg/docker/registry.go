package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nrwiersma/aura/pkg/image"
)

// Registry is a docker registry.
type Registry struct {
	client *docker.Client
}

// NewRegistry returns a registry.
func NewRegistry() (*Registry, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not create docker client: %w", err)
	}

	return &Registry{
		client: client,
	}, nil
}

// Resolve resolves a docker image.
func (r *Registry) Resolve(ctx context.Context, img image.Image) (image.Image, error) {
	opts := docker.PullImageOptions{
		Registry:   img.Registry,
		Repository: img.Repository,
		Tag:        img.Tag,
		Context:    ctx,
	}
	if img.Digest != "" {
		opts.Tag = img.Digest
	}
	err := r.client.PullImage(opts, docker.AuthConfiguration{})
	if err != nil {
		return img, fmt.Errorf("pulling image: %w", err)
	}

	if img.Digest != "" {
		return img, nil
	}

	i, err := r.client.InspectImage(img.String())
	if err != nil {
		return img, fmt.Errorf("inspecting image: %w", err)
	}

	if len(i.RepoDigests) == 0 {
		return img, nil
	}

	return image.Decode(i.RepoDigests[0])
}

// ExtractProcfile extracts a procfile from an image.
func (r *Registry) ExtractProcfile(ctx context.Context, img string) ([]byte, error) {
	ctrID, err := r.createContainer(ctx, img)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.removeContainer(ctx, ctrID) }()

	ctr, err := r.client.InspectContainerWithOptions(docker.InspectContainerOptions{
		ID:      ctrID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	basePath := ""
	if ctr.Config != nil {
		basePath = ctr.Config.WorkingDir
	}
	path := filepath.Join(basePath, "Procfile")

	b, err := r.downloadFile(ctx, ctrID, path)
	if err != nil {
		return nil, fmt.Errorf("downloading procfile: %w", err)
	}
	return b, nil
}

func (r *Registry) createContainer(ctx context.Context, img string) (string, error) {
	ctr, err := r.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: img,
		},
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	return ctr.ID, nil
}

func (r *Registry) removeContainer(ctx context.Context, id string) error {
	if err := r.client.RemoveContainer(docker.RemoveContainerOptions{
		ID:      id,
		Context: ctx,
	}); err != nil {
		return fmt.Errorf("removing container: %w", err)
	}
	return nil
}

func (r *Registry) downloadFile(ctx context.Context, id, path string) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.client.DownloadFromContainer(id, docker.DownloadFromContainerOptions{
		OutputStream: &buf,
		Path:         path,
		Context:      ctx,
	}); err != nil {
		return nil, err
	}

	return readTar(buf.Bytes())
}

func readTar(b []byte) ([]byte, error) {
	r := tar.NewReader(bytes.NewReader(b))

	if _, err := r.Next(); err != nil {
		return nil, fmt.Errorf("reading tar: %w", err)
	}

	var buf bytes.Buffer
	// nolint:gosec // This is fine.
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("reading tar: %w", err)
	}
	return buf.Bytes(), nil
}
