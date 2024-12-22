package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

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

	if f.LeagueUUID != nil {
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

func (d *DB) InsertMatch(ctx context.Context, ec sqlx.ExecerContext, m scouting.Match) error {
	sb := squirrel.Insert("match").SetMap(map[string]any{
		"uuid":            m.UUID,
		"league_uuid":     m.LeagueUUID,
		"away_team_uuid":  m.AwayTeamUUID,
		"home_team_uuid":  m.HomeTeamUUID,
		"created_by":      m.CreatedBy,
		"home_score":      m.HomeScore,
		"away_score":      m.AwayScore,
		"organization_id": m.OrganizationID,
		"starts_at":       m.StartsAt,
		"finished_at":     m.FinishedAt,
		"created_at":      m.CreatedAt,
		"modified_at":     m.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}

func (d *DB) UpdateMatch(ctx context.Context, ec sqlx.ExecerContext, m scouting.Match) error {
	sb := squirrel.Update("match").SetMap(map[string]any{
		"home_score":  m.HomeScore,
		"away_score":  m.AwayScore,
		"finished_at": m.FinishedAt,
		"modified_at": m.ModifiedAt,
	}).Where(squirrel.Eq{
		"uuid": m.UUID,
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

func (d *DB) GetOrganizationLeague(ctx context.Context, qr sqlx.QueryerContext, oid string, uuid uuid.UUID) (scouting.League, error) {
	sb := squirrel.Select(leagueCols()...).From("league AS league").
		InnerJoin("organization_league ON organization_league.league_uuid=league.uuid").
		Where(squirrel.Eq{
			"league.uuid":                         uuid,
			"organization_league.organization_id": oid,
		})

	sq, args := sb.MustSql()

	var l scouting.League

	err := sqlx.GetContext(ctx, qr, &l, sq, args...)
	switch {
	case err == nil:
		return l, nil
	case errors.Is(err, sql.ErrNoRows):
		return scouting.League{}, scouting.ErrStoreNotFound
	default:
		return scouting.League{}, err
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

func (d *DB) DeleteOrganizationLeagues(ctx context.Context, ec sqlx.ExecerContext, oid string) error {
	sb := squirrel.Delete("organization_league").Where(squirrel.Eq{
		"organization_id": oid,
	})

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}

func (d *DB) UpsertAccount(ctx context.Context, ec sqlx.ExecerContext, u scouting.Account) error {
	sb := squirrel.Insert("account").SetMap(map[string]any{
		"id":          u.ID,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"avatar_url":  u.AvatarURL,
		"created_at":  u.CreatedAt,
		"modified_at": u.ModifiedAt,
	}).Suffix("ON CONFLICT (id) DO UPDATE set first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, avatar_url = EXCLUDED.avatar_url, modified_at = EXCLUDED.modified_at")

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}

func (d *DB) UpsertOrganizationAccount(ctx context.Context, ec sqlx.ExecerContext, oid, aid string) error {
	sb := squirrel.Insert("organization_account").SetMap(map[string]any{
		"account_id":      aid,
		"organization_id": oid,
	}).Suffix("ON CONFLICT DO NOTHING")

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}

func matchCols() []string {
	return []string{
		`match.uuid AS "match.uuid"`,
		`match.league_uuid AS "match.league_uuid"`,
		`match.away_team_uuid AS "match.away_team_uuid"`,
		`match.home_team_uuid AS "match.home_team_uuid"`,
		`match.created_by AS "match.created_by"`,
		`match.home_score AS "match.home_score"`,
		`match.away_score AS "match.away_score"`,
		`match.organization_id AS "match.organization_id"`,
		`match.starts_at AS "match.starts_at"`,
		`match.finished_at AS "match.finished_at"`,
		`match.created_at AS "match.created_at"`,
		`match.modified_at AS "match.modified_at"`,
	}
}

func (d *DB) GetOrganizationMatch(ctx context.Context, qr sqlx.QueryerContext, oid string, muuid uuid.UUID, lock bool) (scouting.Match, error) {
	sb := squirrel.Select(matchCols()...).From("match AS match").Where(squirrel.And{
		squirrel.Eq{"uuid": muuid},
		squirrel.Eq{"organization_id": oid},
	})

	if lock {
		sb = sb.Suffix("FOR UPDATE")
	}

	sq, args := sb.MustSql()

	var m scouting.Match

	err := sqlx.GetContext(ctx, qr, &m, sq, args...)
	switch {
	case err == nil:
		return m, nil
	case errors.Is(err, sql.ErrNoRows):
		return scouting.Match{}, scouting.ErrStoreNotFound
	default:
		return scouting.Match{}, err
	}
}

func (d *DB) InsertMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms scouting.MatchScout) error {
	sb := squirrel.Insert("match_scout").SetMap(map[string]any{
		"match_uuid": ms.MatchUUID,
		"account_id": ms.AccountID,
		"mode":       ms.Mode,
		"submode":    ms.Submode,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return err
}

func matchScoutCols() []string {
	return []string{
		`match_scout.match_uuid AS "match_scout.match_uuid"`,
		`match_scout.account_id AS "match_scout.account_id"`,
		`match_scout.mode AS "match_scout.mode"`,
		`match_scout.submode AS "match_scout.submode"`,
		`match_scout.finished_at AS "match_scout.finished_at"`,
	}
}

func (d *DB) SelectMatchScouts(ctx context.Context, qr sqlx.QueryerContext, f scouting.MatchScoutFilter) ([]scouting.MatchScout, error) {
	sb := squirrel.Select(matchScoutCols()...).From("match_scout AS match_scout")

	if f.MatchUUID != nil {
		sb = sb.Where(squirrel.Eq{
			"match_uuid": *f.MatchUUID,
		})
	}

	sql, args := sb.MustSql()

	var smm []scouting.MatchScout

	if err := sqlx.SelectContext(ctx, qr, &smm, sql, args...); err != nil {
		return nil, err
	}

	return smm, nil
}
