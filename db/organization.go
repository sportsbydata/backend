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

func organizationCols() []string {
	return []string{
		`organization.id AS "organization.uuid"`,
		`organization.scouting_config AS "organization.scouting_config"`,
		`organization.created_at AS "organization.created_at"`,
		`organization.modified_at AS "organization.modified_at"`,
	}
}

func (d *DB) SelectOrganizations(ctx context.Context, qr sqlx.QueryerContext, f scouting.OrganizationFilter) ([]scouting.Organization, error) {
	sb := squirrel.Select(organizationCols()...).From("organization AS organization")

	if len(f.IDs) > 0 {
		sb = sb.Where(squirrel.Eq{"id": f.IDs})
	}

	sql, args := sb.MustSql()

	var oo []scouting.Organization

	if err := sqlx.SelectContext(ctx, qr, &oo, sql, args...); err != nil {
		return nil, err
	}

	return oo, nil
}
