package scouting

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type AccountFilter struct {
	OrganizationID *string
}

type Account struct {
	ID         string    `db:"account.id"`
	FirstName  string    `db:"account.first_name"`
	LastName   string    `db:"account.last_name"`
	AvatarURL  string    `db:"account.avatar_url"`
	CreatedAt  time.Time `db:"account.created_at"`
	ModifiedAt time.Time `db:"account.modified_at"`
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
