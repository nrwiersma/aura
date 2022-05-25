package image

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// Image contains the info for a docker image.
type Image struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string
}

// Scan decodes an image from a database field.
func (i *Image) Scan(src any) error {
	b, ok := src.(string)
	if !ok {
		return nil
	}

	img, err := Decode(b)
	if err != nil {
		return err
	}
	*i = img
	return nil
}

// Value encodes an image into a database field.
func (i Image) Value() (driver.Value, error) {
	return driver.Value(i.String()), nil
}

// String returns the string representation of an image.
func (i Image) String() string {
	repo := i.Repository
	if i.Registry != "" {
		repo = i.Registry + "/" + repo
	}

	switch {
	case i.Digest != "":
		return repo + "@" + i.Digest
	case i.Tag != "":
		return repo + ":" + i.Tag
	default:
		return repo
	}
}

// Decode decodes a docker image reference.
func Decode(img string) (Image, error) {
	repo, tag := splitRepositoryTag(img)
	reg, repo := splitRepository(repo)

	if repo == "" {
		return Image{}, errors.New("invalid image format")
	}

	i := Image{
		Registry:   reg,
		Repository: repo,
	}
	if strings.Contains(tag, ":") {
		i.Digest = tag
	} else {
		i.Tag = tag
	}
	return i, nil
}

// Taken from https://github.com/docker/docker/blob/50a1d0f0ef83a9ed55ea2caaa79539ec835877a3/pkg/parsers/parsers.go#L71-L89
func splitRepositoryTag(repos string) (repo, tag string) {
	n := strings.Index(repos, "@")
	if n >= 0 {
		parts := strings.Split(repos, "@")
		return parts[0], parts[1]
	}
	n = strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, ""
}

func splitRepository(fullRepo string) (registry, path string) {
	parts := strings.Split(fullRepo, "/")

	if len(parts) < 2 {
		return "", parts[0]
	}

	if len(parts) == 2 {
		return "", strings.Join(parts, "/")
	}

	return parts[0], strings.Join(parts[1:], "/")
}
