package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

func (d *DB) InsertOrganization(ctx context.Context, ec sqlx.ExecerContext, o scouting.Organization) error {
	sb := squirrel.Insert("organization").SetMap(map[string]any{
		"id":              o.ID,
		"scouting_config": o.ScoutingConfig,
		"created_at":      o.CreatedAt,
		"modified_at":     o.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleError(err)
}
