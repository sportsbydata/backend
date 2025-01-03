package scouting

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Organization struct {
	ID             string         `db:"organization.id"`
	ScoutingConfig ScoutingConfig `db:"organization.scouting_config"`
	CreatedAt      time.Time      `db:"organization.created_at"`
	ModifiedAt     time.Time      `db:"organization.modified_at"`
}

func CreateOrganization(ctx context.Context, sdb *sqlx.DB, store Store, id string) (Organization, error) {
	tnow := time.Now()

	o := Organization{
		ID:             id,
		ScoutingConfig: DefaultScoutingConfig,
		CreatedAt:      tnow,
		ModifiedAt:     tnow,
	}

	if err := store.InsertOrganization(ctx, sdb, o); err != nil {
		slog.Error("inserting organization", slog.Any("error", err))

		return Organization{}, errInternal
	}

	return o, nil
}
