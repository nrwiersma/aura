package image_test

import (
	"database/sql/driver"
	"testing"

	"github.com/nrwiersma/aura/pkg/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImage_Scan(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		want    image.Image
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles valid type",
			in:      "foo/bar:latest",
			want:    image.Image{Repository: "foo/bar", Tag: "latest"},
			wantErr: require.NoError,
		},
		{
			name:    "handles invalid type",
			in:      1,
			wantErr: require.NoError,
		},
		{
			name:    "handles empty image",
			in:      "",
			wantErr: require.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			img := &image.Image{}

			err := img.Scan(test.in)

			test.wantErr(t, err)
			assert.Equal(t, test.want, *img)
		})
	}
}

func TestImage_Value(t *testing.T) {
	img := image.Image{Repository: "foo/bar", Tag: "latest"}

	got, err := img.Value()

	require.NoError(t, err)
	assert.Equal(t, driver.Value("foo/bar:latest"), got)
}

func TestImage_String(t *testing.T) {
	tests := []struct {
		name string
		img  image.Image
		want string
	}{
		{
			name: "handles image with tag",
			img:  image.Image{Registry: "foo", Repository: "bar/baz", Tag: "latest"},
			want: "foo/bar/baz:latest",
		},
		{
			name: "handles image with digest",
			img:  image.Image{Registry: "foo", Repository: "bar/baz", Digest: "sha256:c3ab8"},
			want: "foo/bar/baz@sha256:c3ab8",
		},
		{
			name: "handles image with no tag or digest",
			img:  image.Image{Registry: "foo", Repository: "bar/baz"},
			want: "foo/bar/baz",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := test.img.String()

			assert.Equal(t, test.want, got)
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		img     string
		want    image.Image
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles image with tag",
			img:     "foo/bar:latest",
			want:    image.Image{Repository: "foo/bar", Tag: "latest"},
			wantErr: require.NoError,
		},
		{
			name:    "handles image with digest",
			img:     "foo/bar@sha256:c3ab8",
			want:    image.Image{Repository: "foo/bar", Digest: "sha256:c3ab8"},
			wantErr: require.NoError,
		},
		{
			name:    "handles image with registry",
			img:     "ghcr.io/foo/bar:latest",
			want:    image.Image{Registry: "ghcr.io", Repository: "foo/bar", Tag: "latest"},
			wantErr: require.NoError,
		},
		{
			name:    "handles image with single part repo",
			img:     "bar:latest",
			want:    image.Image{Repository: "bar", Tag: "latest"},
			wantErr: require.NoError,
		},
		{
			name:    "handles image with no tag",
			img:     "foo/bar",
			want:    image.Image{Repository: "foo/bar"},
			wantErr: require.NoError,
		},
		{
			name:    "handles image empty string",
			img:     "",
			wantErr: require.Error,
		},
		{
			name:    "handles image with colon in repo",
			img:     "foo/bar:baz/bat",
			want:    image.Image{Registry: "foo", Repository: "bar:baz/bat"},
			wantErr: require.NoError,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := image.Decode(test.img)

			test.wantErr(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
