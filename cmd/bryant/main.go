package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"
	_ "github.com/jackc/pgx/v5/stdlib" // Standard library bindings for pgxj:w
	"github.com/sportsbydata/backend/db"
	"github.com/sportsbydata/backend/router"
)

var envCfg struct {
	Lambda struct {
		Enabled bool   `env:"ENABLED"`
		DBKey   string `env:"DB_KEY"`
	} `env:"LAMBDA"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	ClerkKey    string `env:"CLERK_KEY"`
	CorsBypass  bool   `env:"CORS_BYPASS"`
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
		EnvPrefix: "BRYANT",
	})

	if err := loader.Load(); err != nil {
		slog.Error("loading config", slog.Any("error", err))

		os.Exit(1)
	}

	if envCfg.Lambda.Enabled {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			slog.Error("loading default config", slog.Any("error", err))

			os.Exit(1)
		}

		client := ssm.NewFromConfig(cfg)

		name := envCfg.Lambda.DBKey
		decrypt := true

		o, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
			Name:           &name,
			WithDecryption: &decrypt,
		})
		if err != nil {
			slog.Error("getting parameter", slog.Any("error", err))

			os.Exit(1)
		}

		envCfg.DatabaseDSN = *o.Parameter.Value
	}

	clerk.SetKey(envCfg.ClerkKey)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	sdb, err := db.Connect(ctx, envCfg.DatabaseDSN)
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))

		os.Exit(1)
	}

	defer sdb.Close()

	if err := db.Migrate(sdb.DB); err != nil {
		slog.Error("migrating", slog.Any("error", err))

		os.Exit(1)
	}

	r := router.New(sdb, envCfg.CorsBypass)

	slog.Info("test")

	if envCfg.Lambda.Enabled {
		lambda.Start(httpadapter.New(r.Handler()).ProxyWithContext)

		return
	}

	s := newServer(envCfg.HTTP.Addr, r.Handler())

	s.run(ctx)
}

type server struct {
	addr string
	h    http.Handler
}

func newServer(addr string, h http.Handler) *server {
	return &server{
		addr: addr,
		h:    h,
	}
}

func (s *server) run(ctx context.Context) {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.h,
	}

	go func() {
		slog.Info("starting server", slog.String("addr", s.addr))

		err := srv.ListenAndServe()
		switch {
		case err == nil:
		case errors.Is(err, http.ErrServerClosed):
			return
		default:
			slog.Error("listening", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	slog.Info("received interr")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := srv.Shutdown(ctx)
	switch {
	case err == nil:
	case errors.Is(err, http.ErrServerClosed):
	default:
		slog.Error("listening", slog.Any("error", err))

		return
	}
}
