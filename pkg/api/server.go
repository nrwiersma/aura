package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hamba/logger/v2"
	mw "github.com/hamba/pkg/v2/http/middleware"
	"github.com/hamba/statter/v2"
	"github.com/nrwiersma/aura/pkg/controller"
)

// Delegate represents an aura delegate.
type Delegate interface {
	App(ctx context.Context, q controller.AppsQuery) (*controller.App, error)
	Apps(ctx context.Context, q controller.AppsQuery) ([]*controller.App, error)
	Create(ctx context.Context, cfg controller.CreateConfig) (*controller.App, error)
	Destroy(ctx context.Context, cfg controller.DestroyConfig) error
	Release(ctx context.Context, q controller.ReleasesQuery) (*controller.Release, error)
	Releases(ctx context.Context, q controller.ReleasesQuery) ([]*controller.Release, error)
	Deploy(ctx context.Context, cfg controller.DeployConfig) (*controller.Release, error)
}

// Server serves api requests.
type Server struct {
	app Delegate

	h http.Handler

	log *logger.Logger
}

// New returns an api server.
func New(app Delegate, log *logger.Logger, stats *statter.Statter) *Server {
	srv := &Server{
		app: app,
		log: log,
	}

	srv.h = srv.routes(stats.With("api"))

	return srv
}

func (s *Server) routes(stats *statter.Statter) http.Handler {
	mux := chi.NewMux()

	mux.Route("/apps", func(r chi.Router) {
		r.With(mw.Stats("get_apps", stats)).Get("/", s.handleGetApps())
		r.With(mw.Stats("get_app", stats)).Get("/{app}", s.handleGetApp())
		r.With(mw.Stats("create_app", stats)).Post("/", s.handleCreateApp())
		r.With(mw.Stats("destroy_app", stats)).Delete("/{app}", s.handleDestroyApp())

		r.With(mw.Stats("get_releases", stats)).Get("/{app}/releases", s.handleGetReleases())
		r.With(mw.Stats("get_release", stats)).Get("/{app}/releases/{version}", s.handleGetRelease())
		r.With(mw.Stats("deploy_app", stats)).Post("/{app}/deploys", s.handlerDeployApp())
	})

	return mux
}

// ServeHTTP serves an HTTP request.
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.h.ServeHTTP(rw, req)
}
