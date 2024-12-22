package scouting

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Account struct {
	ID         string    `db:"user.id" json:"id"`
	FirstName  string    `db:"first_name" json:"first_name"`
	LastName   string    `db:"last_name" json:"last_name"`
	AvatarURL  string    `db:"avatar_url" json:"avatar_url"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	ModifiedAt time.Time `db:"modified_at" json:"modified_at"`
}

func UpsertAccount(ctx context.Context, sdb *sqlx.DB, store Store, oid string, a Account) error {
	logger := slog.With(slog.String("organization_id", oid), slog.String("user_id", a.ID))

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning tx", slog.Any("error", err))

		return errInternal
	}

	defer tx.Rollback()

	if err = store.UpsertAccount(ctx, tx, a); err != nil {
		logger.Error("upserting account", slog.Any("error", err))

		return errInternal
	}

	if err = store.UpsertOrganizationAccount(ctx, tx, oid, a.ID); err != nil {
		logger.Error("upserting account organization", slog.Any("error", err))

		return errInternal
	}

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return errInternal
	}

	return nil
}
