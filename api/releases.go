package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nrwiersma/aura"
	"github.com/nrwiersma/aura/pkg/image"
	"github.com/nrwiersma/aura/pkg/render"
)

type releaseResp struct {
	ID        string     `json:"id"`
	App       appResp    `json:"app,omitempty"`
	Image     string     `json:"image"`
	Version   int        `json:"version"`
	Procfile  string     `json:"procfile"`
	CreatedAt *time.Time `json:"createdAt"`
}

func toReleaseResp(release *aura.Release) releaseResp {
	resp := releaseResp{
		ID:        release.ID,
		Version:   release.Version,
		Procfile:  string(release.Procfile),
		CreatedAt: release.CreatedAt,
	}
	if release.App != nil {
		resp.App = toAppResp(release.App)
	}
	if release.Image != nil {
		resp.Image = release.Image.String()
	}
	return resp
}

func (s *Server) handleGetReleases() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		appID := chi.URLParam(req, "app")

		log := s.log.With(lctx.Str("app_id", appID))

		app, err := s.app.AppsFind(req.Context(), aura.AppsQuery{ID: appID})
		if err != nil {
			switch {
			case errors.Is(err, aura.ErrNotFound):
				log.Debug("App not found")
				render.JSONError(rw, http.StatusNotFound, "app not found")
			default:
				s.log.Error("Could not deploy app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		releases, err := s.app.Releases(req.Context(), aura.ReleasesQuery{App: app})
		if err != nil {
			log.Error("Could not get releases", lctx.Error("error", err))
			render.JSONInternalServerError(rw)
			return
		}

		resp := make([]releaseResp, 0, len(releases))
		for _, release := range releases {
			resp = append(resp, toReleaseResp(release))
		}
		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) handleGetRelease() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		appID := chi.URLParam(req, "app")

		verStr := chi.URLParam(req, "version")
		ver, err := strconv.Atoi(verStr)
		if err != nil {
			s.log.Debug("Could not convert version to int", lctx.Error("error", err))
			render.JSONError(rw, http.StatusBadRequest, "version must be a positive integer")
			return
		}

		log := s.log.With(lctx.Str("app_id", appID), lctx.Int("version", ver))

		app, err := s.app.AppsFind(req.Context(), aura.AppsQuery{ID: appID})
		if err != nil {
			switch {
			case errors.Is(err, aura.ErrNotFound):
				log.Debug("App not found")
				render.JSONError(rw, http.StatusNotFound, "app not found")
			default:
				log.Error("Could not get app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		release, err := s.app.ReleasesFind(req.Context(), aura.ReleasesQuery{App: app, Version: ver})
		if err != nil {
			switch {
			case errors.Is(err, aura.ErrNotFound):
				log.Debug("Release not found")
				render.JSONError(rw, http.StatusNotFound, "release not found")
			default:
				log.Error("Could not get release", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		resp := toReleaseResp(release)
		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) handlerDeployApp() http.HandlerFunc {
	type deployAppReq struct {
		Image string `json:"image"`
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		appID := chi.URLParam(req, "app")

		log := s.log.With(lctx.Str("app_id", appID))

		var appReq deployAppReq
		if err := json.NewDecoder(req.Body).Decode(&appReq); err != nil {
			log.Debug("Could not unmarshal body", lctx.Error("error", err))
			render.JSONError(rw, http.StatusBadRequest, "invalid app deployment data")
			return
		}

		img, err := image.Decode(appReq.Image)
		if err != nil {
			s.log.Debug("Invalid deployment", lctx.Error("error", err))
			render.JSONErrorf(rw, http.StatusBadRequest, "invalid app deploy: %v", err)
			return
		}

		resp, err := s.deployApp(req.Context(), appID, img)
		if err != nil {
			switch {
			case errors.Is(err, aura.ErrNotFound):
				log.Debug("App not found")
				render.JSONError(rw, http.StatusNotFound, "app not found")
			default:
				s.log.Error("Could not deploy app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) deployApp(ctx context.Context, appID string, img image.Image) (releaseResp, error) {
	app, err := s.app.AppsFind(ctx, aura.AppsQuery{ID: appID})
	if err != nil {
		return releaseResp{}, err
	}

	release, err := s.app.Deploy(ctx, aura.DeployConfig{
		App:   app,
		Image: img,
	})
	if err != nil {
		return releaseResp{}, err
	}

	return toReleaseResp(release), nil
}
