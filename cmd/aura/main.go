package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v2"
)

const (
	flagAddr          = "addr"
	flagDBDSN         = "db.dsn"
	flagDBAutoMigrate = "db.auto-migrate"
)

var version = "¯\\_(ツ)_/¯"

var flags = cmd.Flags{
	&cli.StringFlag{
		Name:    flagAddr,
		Usage:   "The address to listen on",
		Value:   ":8080",
		EnvVars: []string{strcase.ToSNAKE(flagAddr)},
	},
	&cli.StringFlag{
		Name:     flagDBDSN,
		Usage:    "The DSN to connect to the database",
		Required: true,
		EnvVars:  []string{strcase.ToSNAKE(flagDBDSN)},
	},
	&cli.BoolFlag{
		Name:    flagDBAutoMigrate,
		Usage:   "Determines of the database should run migrations on startup",
		Value:   true,
		EnvVars: []string{strcase.ToSNAKE(flagDBAutoMigrate)},
	},
}.Merge(cmd.LogFlags, cmd.StatsFlags)

func main() {
	os.Exit(realMain(os.Args))
}

func realMain(args []string) (code int) {
	defer func() {
		if v := recover(); v != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Panic: %v\n%s\n", v, debug.Stack())
			code = 1
		}
	}()
	app := cli.NewApp()
	app.Name = "aura"
	app.Usage = "Demo application"
	app.Version = version
	app.Flags = flags
	app.Action = runServer

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.RunContext(ctx, args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return 1
	}
	return 0
}
