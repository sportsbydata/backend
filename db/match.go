package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

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

func (d *DB) SelectMatches(ctx context.Context, qr sqlx.QueryerContext, f scouting.MatchFilter, lock bool) ([]scouting.Match, error) {
	var dec squirrel.And

	if f.OrganizationID != "" {
		dec = append(dec, squirrel.Eq{"match.organization_id": f.OrganizationID})
	}

	if f.Active {
		dec = append(dec, squirrel.Expr("match.finished_at IS NULL"))
	} else {
		dec = append(dec, squirrel.Expr("match.finished_at IS NOT NULL"))
	}

	if !f.UUID.IsNil() {
		dec = append(dec, squirrel.Eq{"uuid": f.UUID})
	}

	sb := squirrel.Select(matchCols()...).From("match AS match")

	if len(dec) > 0 {
		sb = sb.Where(dec)
	}

	if lock {
		sb = sb.Suffix("FOR UPDATE")
	}

	sql, args := sb.MustSql()

	var mm []scouting.Match

	if err := sqlx.SelectContext(ctx, qr, &mm, sql, args...); err != nil {
		return nil, err
	}

	return mm, nil
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

func (d *DB) UpdateMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms scouting.MatchScout) error {
	sb := squirrel.Update("match_scout").SetMap(map[string]any{
		"finished_at": ms.FinishedAt,
	}).Where(squirrel.And{
		squirrel.Eq{"account_id": ms.AccountID},
		squirrel.Eq{"match_uuid": ms.MatchUUID},
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

	var dec squirrel.And

	if f.MatchUUID != nil {
		dec = append(dec, squirrel.Eq{
			"match_scout.match_uuid": *f.MatchUUID,
		})
	}

	if f.MatchOrganizationID != nil {
		sb = sb.InnerJoin("match ON match.uuid=match_scout.match_uuid")

		dec = append(dec, squirrel.Eq{
			"match.organization_id": *f.MatchOrganizationID,
		})
	}

	if len(dec) > 0 {
		sb = sb.Where(dec)
	}

	sql, args := sb.MustSql()

	var smm []scouting.MatchScout

	if err := sqlx.SelectContext(ctx, qr, &smm, sql, args...); err != nil {
		return nil, err
	}

	return smm, nil
}
