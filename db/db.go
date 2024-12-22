package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"net/http"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // for migration support
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jmoiron/sqlx"
)

type DB struct{}

//go:embed migrations
var migrations embed.FS

func Connect(ctx context.Context, dsn string) (*sqlx.DB, error) {
	sdb, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = sdb.Ping(); err != nil {
		return nil, err
	}

	squirrel.StatementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return sdb, nil
}

func Migrate(pdb *sql.DB) error {
	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(pdb, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("source", src, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	switch {
	case err == nil:
	case errors.Is(err, migrate.ErrNoChange):
	default:
		return err
	}

	return nil
}
