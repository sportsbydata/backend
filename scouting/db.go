package scouting

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"net/http"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/sbd"
)

//go:embed migrations
var migrations embed.FS

func ConnectDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	sdb, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, handleDbError(err)
	}

	if err = sdb.Ping(); err != nil {
		return nil, handleDbError(err)
	}

	squirrel.StatementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return sdb, nil
}

func Migrate(pdb *sql.DB) error {
	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return handleDbError(err)
	}

	driver, err := postgres.WithInstance(pdb, &postgres.Config{})
	if err != nil {
		return handleDbError(err)
	}

	m, err := migrate.NewWithInstance("source", src, "postgres", driver)
	if err != nil {
		return handleDbError(err)
	}

	err = m.Up()
	switch {
	case err == nil, errors.Is(err, migrate.ErrNoChange):
		return nil
	default:
		return handleDbError(err)
	}
}

func accountCols() []string {
	return []string{
		`account.id AS "account.id"`,
		`account.first_name AS "account.first_name"`,
		`account.last_name AS "account.last_name"`,
		`account.avatar_url AS "account.avatar_url"`,
		`account.created_at AS "account.created_at"`,
		`account.modified_at AS "account.modified_at"`,
	}
}

func SelectAccounts(ctx context.Context, qr sqlx.QueryerContext, f AccountFilter) ([]Account, error) {
	sb := squirrel.Select(accountCols()...).From("account AS account")

	if f.OrganizationID != "" {
		sb = sb.InnerJoin("organization_account ON organization_account.account_id=account.id").
			Where(squirrel.Eq{"organization_account.organization_id": &f.OrganizationID})
	}

	sql, args := sb.MustSql()

	var aa []Account

	if err := sqlx.SelectContext(ctx, qr, &aa, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return aa, nil
}

func insertAccount(ctx context.Context, ec sqlx.ExecerContext, u Account) error {
	sb := squirrel.Insert("account").SetMap(map[string]any{
		"id":          u.ID,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"avatar_url":  u.AvatarURL,
		"created_at":  u.CreatedAt,
		"modified_at": u.ModifiedAt,
	})

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return handleDbError(err)
}

func upsertOrganizationAccount(ctx context.Context, ec sqlx.ExecerContext, oid, aid string) error {
	sb := squirrel.Insert("organization_account").SetMap(map[string]any{
		"account_id":      aid,
		"organization_id": oid,
	}).Suffix("ON CONFLICT DO NOTHING")

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return handleDbError(err)
}

func insertLeague(ctx context.Context, ec sqlx.ExecerContext, l League) error {
	sb := squirrel.Insert("league").SetMap(map[string]any{
		"uuid":        l.UUID,
		"name":        l.Name,
		"created_at":  l.CreatedAt,
		"modified_at": l.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func insertLeagueTeam(ctx context.Context, ec sqlx.ExecerContext, luuid, tuuid uuid.UUID) error {
	sb := squirrel.Insert("league_team").SetMap(map[string]any{
		"league_uuid": luuid,
		"team_uuid":   tuuid,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func leagueCols() []string {
	return []string{
		`league.uuid AS "league.uuid"`,
		`league.name AS "league.name"`,
		`league.created_at AS "league.created_at"`,
		`league.modified_at AS "league.modified_at"`,
	}
}

func insertOrganizationLeague(ctx context.Context, ec sqlx.ExecerContext, oid string, luuid uuid.UUID) error {
	sb := squirrel.Insert("organization_league").SetMap(map[string]any{
		"organization_id": oid,
		"league_uuid":     luuid,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func SelectLeagues(ctx context.Context, qr sqlx.QueryerContext, f LeagueFilter) ([]League, error) {
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

	var ll []League

	if err := sqlx.SelectContext(ctx, qr, &ll, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return ll, nil
}

func deleteOrganizationLeagues(ctx context.Context, ec sqlx.ExecerContext, oid string) error {
	sb := squirrel.Delete("organization_league").Where(squirrel.Eq{
		"organization_id": oid,
	})

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return handleDbError(err)
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

func SelectMatches(ctx context.Context, qr sqlx.QueryerContext, f MatchFilter, lock bool) ([]Match, error) {
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

	var mm []Match

	if err := sqlx.SelectContext(ctx, qr, &mm, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return mm, nil
}

func insertMatch(ctx context.Context, ec sqlx.ExecerContext, m Match) error {
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
	return handleDbError(err)
}

func updateMatch(ctx context.Context, ec sqlx.ExecerContext, m Match) error {
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
	return handleDbError(err)
}

func insertMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms MatchScout) error {
	sb := squirrel.Insert("match_scout").SetMap(map[string]any{
		"match_uuid": ms.MatchUUID,
		"account_id": ms.AccountID,
		"mode":       ms.Mode,
		"submode":    ms.Submode,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func updateMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms MatchScout) error {
	sb := squirrel.Update("match_scout").SetMap(map[string]any{
		"finished_at": ms.FinishedAt,
	}).Where(squirrel.And{
		squirrel.Eq{"account_id": ms.AccountID},
		squirrel.Eq{"match_uuid": ms.MatchUUID},
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
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

func SelectMatchScouts(ctx context.Context, qr sqlx.QueryerContext, f MatchScoutFilter) ([]MatchScout, error) {
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

	var smm []MatchScout

	if err := sqlx.SelectContext(ctx, qr, &smm, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return smm, nil
}

func insertOrganization(ctx context.Context, ec sqlx.ExecerContext, o Organization) error {
	sb := squirrel.Insert("organization").SetMap(map[string]any{
		"id":              o.ID,
		"scouting_config": o.ScoutingConfig,
		"created_at":      o.CreatedAt,
		"modified_at":     o.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func organizationCols() []string {
	return []string{
		`organization.id AS "organization.uuid"`,
		`organization.scouting_config AS "organization.scouting_config"`,
		`organization.created_at AS "organization.created_at"`,
		`organization.modified_at AS "organization.modified_at"`,
	}
}

func selectOrganizations(ctx context.Context, qr sqlx.QueryerContext, f OrganizationFilter) ([]Organization, error) {
	sb := squirrel.Select(organizationCols()...).From("organization AS organization")

	if len(f.IDs) > 0 {
		sb = sb.Where(squirrel.Eq{"id": f.IDs})
	}

	sql, args := sb.MustSql()

	var oo []Organization

	if err := sqlx.SelectContext(ctx, qr, &oo, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return oo, nil
}

func teamCols() []string {
	return []string{
		`team.uuid AS "team.uuid"`,
		`team.name AS "team.name"`,
		`team.created_at AS "team.created_at"`,
		`team.modified_at AS "team.modified_at"`,
	}
}

func SelectTeams(ctx context.Context, qr sqlx.QueryerContext, f TeamFilter) ([]Team, error) {
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

	var tt []Team

	if err := sqlx.SelectContext(ctx, qr, &tt, sql, args...); err != nil {
		return nil, handleDbError(err)
	}

	return tt, nil
}

func insertTeam(ctx context.Context, ec sqlx.ExecerContext, t Team) error {
	sb := squirrel.Insert("team").SetMap(map[string]any{
		"uuid":        t.UUID,
		"name":        t.Name,
		"created_at":  t.CreatedAt,
		"modified_at": t.ModifiedAt,
	})

	sql, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sql, args...)
	return handleDbError(err)
}

func handleDbError(err error) error {
	var pge *pgconn.PgError

	if errors.As(err, &pge) && pge.Code == pgerrcode.UniqueViolation {
		return sbd.ErrAlreadyExists
	}

	return err
}
