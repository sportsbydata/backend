package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/cristalhq/aconfig"
	"github.com/jackc/pgx/v5"
)

type config struct {
	DBDSN string `env:"DB_DSN"`
}

func main() {
	var cfg config

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		// feel free to skip some steps :)
		// SkipDefaults: true,
		// SkipFiles:    true,
		// SkipEnv:      true,
		// SkipFlags:    true,
		EnvPrefix: "BRYANT",
	})

	if err := loader.Load(); err != nil {
		slog.Error("loading config", slog.Any("error", err))

		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DBDSN)
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))

		os.Exit(1)
	}

	defer conn.Close(context.Background())

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello World:)"))
	})

	slog.Info("starting server")

	http.ListenAndServe(":8080", nil)
}
