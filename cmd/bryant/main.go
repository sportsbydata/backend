package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/sportsbydata/backend/scouting"
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
	PrometheusBasicAuth []byte `env:"PROMETHEUS_BASIC_AUTH" default:""`
}

func main() {
	if err := run(); err != nil {
		slog.Error("running", slog.Any("error", err))

		os.Exit(1)
	}
}

func run() error {
	loader := aconfig.LoaderFor(&envCfg, aconfig.Config{
		Files: []string{"/etc/bryant/config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
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

	slog.Info("starting")

	clerk.SetKey(envCfg.ClerkKey)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	sdb, err := scouting.ConnectDB(ctx, envCfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer func() {
		if err := sdb.Close(); err != nil {
			slog.Error("closing database", slog.Any("error", err))
		}
	}()

	if err := scouting.Migrate(sdb.DB); err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}

	s := server.New(sdb, envCfg.HTTP.Addr, envCfg.PrometheusBasicAuth, envCfg.Dev)

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
