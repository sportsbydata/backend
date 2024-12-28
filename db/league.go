package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

func (d *DB) InsertLeague(ctx context.Context, ec sqlx.ExecerContext, l scouting.League) error {
	sb := squirrel.Insert("league").SetMap(map[string]any{
		"uuid":        l.UUID,
		"name":        l.Name,
		"created_at":  l.CreatedAt,
		"modified_at": l.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}

func (d *DB) InsertLeagueTeam(ctx context.Context, ec sqlx.ExecerContext, luuid, tuuid uuid.UUID) error {
	sb := squirrel.Insert("league_team").SetMap(map[string]any{
		"league_uuid": luuid,
		"team_uuid":   tuuid,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}

func leagueCols() []string {
	return []string{
		`league.uuid AS "league.uuid"`,
		`league.name AS "league.name"`,
		`league.created_at AS "league.created_at"`,
		`league.modified_at AS "league.modified_at"`,
	}
}

func (d *DB) InsertOrganizationLeague(ctx context.Context, ec sqlx.ExecerContext, oid string, luuid uuid.UUID) error {
	sb := squirrel.Insert("organization_league").SetMap(map[string]any{
		"organization_id": oid,
		"league_uuid":     luuid,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}

func (d *DB) SelectLeagues(ctx context.Context, qr sqlx.QueryerContext, f scouting.LeagueFilter) ([]scouting.League, error) {
	sb := squirrel.Select(leagueCols()...).From("league AS league")

	var dec squirrel.And

	if !f.LeagueUUID.IsNil() {
		dec = append(dec, squirrel.Eq{
			"league.uuid": f.LeagueUUID,
		})
	}

	if f.OrganizationID != "" {
		sb = sb.InnerJoin("organization_league ON organization_league.league_uuid=league.uuid")

		dec = append(dec, squirrel.Eq{
			"organization_league.organization_id": f.OrganizationID,
		})
	}

	if len(dec) > 0 {
		sb = sb.Where(dec)
	}

	sql, args := sb.MustSql()

	var ll []scouting.League

	if err := sqlx.SelectContext(ctx, qr, &ll, sql, args...); err != nil {
		return nil, err
	}

	return ll, nil
}

func (d *DB) DeleteOrganizationLeagues(ctx context.Context, ec sqlx.ExecerContext, oid string) error {
	sb := squirrel.Delete("organization_league").Where(squirrel.Eq{
		"organization_id": oid,
	})

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}
