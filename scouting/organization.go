package scouting

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/sbd"
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

func CreateOrganization(ctx context.Context, sdb *sqlx.DB, id string) (Organization, error) {
	tnow := time.Now()

	o := Organization{
		ID:             id,
		ScoutingConfig: DefaultScoutingConfig,
		CreatedAt:      tnow,
		ModifiedAt:     tnow,
	}

	err := insertOrganization(ctx, sdb, o)
	switch {
	case err == nil:
		// OK.
	case errors.Is(err, sbd.ErrAlreadyExists):
		return Organization{}, err
	default:
		slog.Error("inserting organization", slog.Any("error", err))

		return Organization{}, errInternal
	}

	return o, nil
}

func SelectOrganizations(ctx context.Context, sdb *sqlx.DB, f OrganizationFilter) ([]Organization, error) {
	return selectOrganizations(ctx, sdb, f)
}
