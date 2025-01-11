package scouting

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Organization struct {
	ID             string         `db:"organization.id"`
	ScoutingConfig ScoutingConfig `db:"organization.scouting_config"`
	Name           string         `db:"organizations.name"`
	CreatedAt      time.Time      `db:"organization.created_at"`
	ModifiedAt     time.Time      `db:"organization.modified_at"`
}

type NewOrganization struct {
	ID string
}

type OrganizationFilter struct {
	IDs []string
}

func CreateOrganization(ctx context.Context, sdb *sqlx.DB, store Store, id string) (Organization, error) {
	tnow := time.Now()

	o := Organization{
		ID:             id,
		ScoutingConfig: DefaultScoutingConfig,
		CreatedAt:      tnow,
		ModifiedAt:     tnow,
	}

	err := store.InsertOrganization(ctx, sdb, o)
	switch {
	case err == nil:
		// OK.
	case errors.Is(err, ErrAlreadyExists):
		return Organization{}, err
	default:
		slog.Error("inserting organization", slog.Any("error", err))

		return Organization{}, errInternal
	}

	return o, nil
}

func SelectOrganizations(ctx context.Context, sdb *sqlx.DB, store Store, f OrganizationFilter) ([]Organization, error) {
	return store.SelectOrganizations(ctx, sdb, f)
}
