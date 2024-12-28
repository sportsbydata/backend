package main

import (
	"context"
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
}

func main() {
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
		slog.Error("loading config", slog.Any("error", err))

		os.Exit(1)
	}

	clerk.SetKey(envCfg.ClerkKey)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	sdb, err := db.Connect(ctx, envCfg.DatabaseDSN)
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))

		os.Exit(1)
	}

	defer func() {
		if err := sdb.Close(); err != nil {
			slog.Error("closing database", slog.Any("error", err))
		}
	}()

	if err := db.Migrate(sdb.DB); err != nil {
		slog.Error("migrating", slog.Any("error", err))

		os.Exit(1)
	}

	s := server.New(sdb, &db.DB{}, envCfg.HTTP.Addr)

	s.Run()

	<-ctx.Done()

	shutdownTimeout, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := s.Close(shutdownTimeout); err != nil {
		slog.Error("closing server", slog.Any("error", err))
	}
}
