package scouting

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/sbd"
)

type AccountFilter struct {
	OrganizationID string
	ID             string
}

type Account struct {
	ID         string    `db:"account.id"`
	FirstName  string    `db:"account.first_name"`
	LastName   string    `db:"account.last_name"`
	AvatarURL  string    `db:"account.avatar_url"`
	CreatedAt  time.Time `db:"account.created_at"`
	ModifiedAt time.Time `db:"account.modified_at"`
}

func OnboardAccount(ctx context.Context, sdb *sqlx.DB, oid string, cu *clerk.User) (Account, error) {
	logger := slog.With(slog.String("organization_id", oid), slog.String("user_id", cu.ID))

	if cu.FirstName == nil {
		return Account{}, sbd.NewValidationError("missing first name in clerk")
	}

	if cu.LastName == nil {
		return Account{}, sbd.NewValidationError("missing last name in clerk")
	}

	if cu.ImageURL == nil {
		return Account{}, sbd.NewValidationError("missing image url in clerk")
	}

	tnow := time.Now()

	a := Account{
		ID:         cu.ID,
		FirstName:  *cu.FirstName,
		LastName:   *cu.LastName,
		AvatarURL:  *cu.ImageURL,
		CreatedAt:  tnow,
		ModifiedAt: tnow,
	}

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning tx", slog.Any("error", err))

		return Account{}, errInternal
	}

	defer tx.Rollback()

	err = insertAccount(ctx, tx, a)
	switch {
	case err == nil:
		// OK.
	case errors.Is(err, sbd.ErrAlreadyExists):
		return Account{}, err
	default:
		logger.Error("upserting account", slog.Any("error", err))

		return Account{}, errInternal
	}

	if err = upsertOrganizationAccount(ctx, tx, oid, a.ID); err != nil {
		logger.Error("upserting account organization", slog.Any("error", err))

		return Account{}, errInternal
	}

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return Account{}, errInternal
	}

	return a, nil
}
