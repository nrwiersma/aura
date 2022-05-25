package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nrwiersma/aura/pkg/controller"
	"github.com/nrwiersma/aura/pkg/render"
)

type appResp struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

func toAppResp(app *controller.App) appResp {
	return appResp{
		ID:        app.ID,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		DeletedAt: app.DeletedAt,
	}
}

func (s *Server) handleGetApps() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		apps, err := s.app.Apps(req.Context(), controller.AppsQuery{})
		if err != nil {
			s.log.Error("Could not get apps", lctx.Error("error", err))
			render.JSONInternalServerError(rw)
			return
		}

		resp := make([]appResp, 0, len(apps))
		for _, app := range apps {
			resp = append(resp, toAppResp(app))
		}
		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) handleGetApp() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		appID := chi.URLParam(req, "app")

		log := s.log.With(lctx.Str("app_id", appID))

		app, err := s.app.App(req.Context(), controller.AppsQuery{ID: appID})
		if err != nil {
			switch {
			case errors.Is(err, controller.ErrNotFound):
				log.Debug("App not found")
				render.JSONError(rw, http.StatusNotFound, "app not found")
			default:
				log.Error("Could not get app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		resp := toAppResp(app)
		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) handleCreateApp() http.HandlerFunc {
	type createAppReq struct {
		Name string `json:"name"`
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		var appReq createAppReq
		if err := json.NewDecoder(req.Body).Decode(&appReq); err != nil {
			s.log.Debug("Could not unmarshal body", lctx.Error("error", err))
			render.JSONError(rw, http.StatusBadRequest, "invalid app data")
			return
		}

		app, err := s.app.Create(req.Context(), controller.CreateConfig{Name: appReq.Name})
		if err != nil {
			switch {
			case errors.As(err, &controller.ValidationError{}):
				s.log.Debug("Invalid app", lctx.Error("error", err))
				render.JSONErrorf(rw, http.StatusBadRequest, "invalid app: %v", err)
			default:
				s.log.Error("Could not create app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		resp := toAppResp(app)
		if err = render.JSON(rw, http.StatusOK, resp); err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
	}
}

func (s *Server) handleDestroyApp() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		appID := chi.URLParam(req, "app")

		log := s.log.With(lctx.Str("app_id", appID))

		if err := s.destroyApp(req.Context(), appID); err != nil {
			switch {
			case errors.Is(err, controller.ErrNotFound):
				log.Debug("App not found")
				render.JSONError(rw, http.StatusNotFound, "app not found")
			default:
				log.Error("Could not destroy app", lctx.Error("error", err))
				render.JSONInternalServerError(rw)
			}
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) destroyApp(ctx context.Context, appID string) error {
	app, err := s.app.App(ctx, controller.AppsQuery{ID: appID})
	if err != nil {
		return err
	}

	return s.app.Destroy(ctx, controller.DestroyConfig{App: app})
}
