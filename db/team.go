package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

func teamCols() []string {
	return []string{
		`team.uuid AS "team.uuid"`,
		`team.name AS "team.name"`,
		`team.created_at AS "team.created_at"`,
		`team.modified_at AS "team.modified_at"`,
	}
}

func (d *DB) SelectTeams(ctx context.Context, qr sqlx.QueryerContext, f scouting.TeamFilter) ([]scouting.Team, error) {
	sb := squirrel.Select(teamCols()...).From("team AS team")

	if !f.LeagueUUID.IsNil() {
		sb = sb.InnerJoin(
			"league_team ON league_team.team_uuid=team.uuid").
			Where(squirrel.Eq{"league_uuid": f.LeagueUUID})
	}

	if len(f.UUIDs) > 0 {
		sb = sb.Where(squirrel.Eq{"uuid": f.UUIDs})
	}

	sql, args := sb.MustSql()

	var tt []scouting.Team

	if err := sqlx.SelectContext(ctx, qr, &tt, sql, args...); err != nil {
		return nil, err
	}

	return tt, nil
}

func (d *DB) InsertTeam(ctx context.Context, ec sqlx.ExecerContext, t scouting.Team) error {
	sb := squirrel.Insert("team").SetMap(map[string]any{
		"uuid":        t.UUID,
		"name":        t.Name,
		"created_at":  t.CreatedAt,
		"modified_at": t.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}
