package main

import (
	"net/http"
	"time"

	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	httpx "github.com/hamba/pkg/v2/http"
	mw "github.com/hamba/pkg/v2/http/middleware"
	"github.com/hamba/statter/v2/runtime"
	"github.com/nrwiersma/aura"
	"github.com/nrwiersma/aura/api"
	"github.com/nrwiersma/aura/docker"
	"github.com/urfave/cli/v2"
)

func runServer(c *cli.Context) error {
	ctx := c.Context

	log, err := cmd.NewLogger(c)
	if err != nil {
		return err
	}
	log = log.With(lctx.Str("app", "aura"))

	stats, err := cmd.NewStatter(c, log)
	if err != nil {
		return err
	}
	defer func() { _ = stats.Close() }()
	go runtime.Collect(stats)

	db, err := aura.OpenDB(c.String(flagDBDSN), log)
	if err != nil {
		return err
	}

	if c.Bool(flagDBAutoMigrate) {
		if err = db.Migrate(); err != nil {
			return err
		}
	}

	reg, err := docker.NewRegistry()
	if err != nil {
		return err
	}

	app := aura.New(db, reg)

	apiSrv := api.New(app, log, stats)

	mux := http.NewServeMux()
	mux.Handle("/readyz", httpx.OKHandler())
	mux.Handle("/healthz", httpx.NewHealthHandler(db))
	mux.Handle("/", apiSrv)

	h := mw.WithRecovery(mux, log)

	addr := c.String(flagAddr)
	srv := httpx.NewServer(ctx, addr, h, httpx.WithH2C())

	log.Info("Starting server", lctx.Str("addr", addr))
	srv.Serve(func(err error) {
		log.Error("Server error", lctx.Error("error", err))
	})
	defer func() { _ = srv.Close() }()

	<-c.Done()

	log.Info("Shutting down")
	if err = srv.Shutdown(10 * time.Second); err != nil {
		log.Error("Failed to shutdown server", lctx.Error("error", err))
	}

	return nil
}
