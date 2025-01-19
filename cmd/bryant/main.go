package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sportsbydata/backend/db"
	"github.com/sportsbydata/backend/server"
)

var envCfg struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
	ClerkKey    string `env:"CLERK_KEY"`
	HTTP        struct {
		Addr string `env:"ADDR" default:":8043"`
	} `env:"HTTP"`
	Dev bool `env:"DEV"`
	Log struct {
		JSON bool `env:"JSON"`
	} `env:"LOG"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("running", slog.Any("error", err))

		os.Exit(1)
	}
}

func run() error {
	loader := aconfig.LoaderFor(&envCfg, aconfig.Config{
		Files: []string{"config.toml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".toml": aconfigtoml.New(),
		},
		EnvPrefix:          "BRYANT",
		AllowUnknownFields: true,
		AllowUnknownEnvs:   true,
	})

	if err := loader.Load(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if envCfg.Log.JSON {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		slog.SetDefault(logger)
	}

	clerk.SetKey(envCfg.ClerkKey)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	sdb, err := db.Connect(ctx, envCfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer func() {
		if err := sdb.Close(); err != nil {
			slog.Error("closing database", slog.Any("error", err))
		}
	}()

	if err := db.Migrate(sdb.DB); err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}

	s := server.New(sdb, &db.DB{}, envCfg.HTTP.Addr, envCfg.Dev)

	s.Run()

	<-ctx.Done()

	slog.Info("received interrupt")

	shutdownTimeout, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := s.Close(shutdownTimeout); err != nil {
		slog.Error("closing server", slog.Any("error", err))
	}

	return nil
}
