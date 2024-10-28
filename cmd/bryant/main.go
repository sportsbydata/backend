package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cristalhq/aconfig"
	"github.com/jackc/pgx/v5"
)

var config struct {
	DBDSN string `env:"DB_DSN"`
}

func init() {
	loader := aconfig.LoaderFor(&config, aconfig.Config{
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

	slog.Info("connected")
}

func handle(ctx context.Context, _ json.RawMessage) error {
	conn, err := pgx.Connect(ctx, config.DBDSN)
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))

		os.Exit(1)
	}

	defer conn.Close(context.Background())

	return nil
}

func main() {
	lambda.Start(handle)
}
