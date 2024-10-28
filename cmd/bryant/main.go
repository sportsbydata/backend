package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/cristalhq/aconfig"
	"github.com/jackc/pgx/v5"
)

var envCfg struct {
	DBDSN string `env:"DB_DSN"`
}

var dbDSN string

func init() {
	loader := aconfig.LoaderFor(&envCfg, aconfig.Config{
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

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("loading default config", slog.Any("error", err))

		os.Exit(1)
	}

	client := ssm.NewFromConfig(cfg)

	name := "/sbd/dev/scoutapi/db-dsn"
	decrypt := true

	o, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &decrypt,
	})
	if err != nil {
		slog.Error("getting parameter", slog.Any("error", err))

		os.Exit(1)
	}

	dbDSN = *o.Parameter.Value
}

func handle(ctx context.Context, _ json.RawMessage) error {
	conn, err := pgx.Connect(ctx, dbDSN)
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))

		os.Exit(1)
	}

	defer conn.Close(context.Background())

	slog.Info("connected")

	return nil
}

func main() {
	lambda.Start(handle)
}
