package scouting

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

var (
	ErrStoreNotFound = errors.New("not found")
)

type Store interface {
	InsertOrganization(ctx context.Context, ec sqlx.ExecerContext, o Organization) error
	SelectTeams(ctx context.Context, q sqlx.QueryerContext, f TeamFilter) ([]Team, error)
	InsertTeam(ctx context.Context, ec sqlx.ExecerContext, t Team) error

	DeleteOrganizationLeagues(ctx context.Context, ec sqlx.ExecerContext, oid string) error
	InsertLeague(ctx context.Context, ec sqlx.ExecerContext, l League) error
	InsertLeagueTeam(ctx context.Context, ec sqlx.ExecerContext, luuid, tuuid uuid.UUID) error
	InsertOrganizationLeague(ctx context.Context, ec sqlx.ExecerContext, oid string, luuid uuid.UUID) error
	SelectLeagues(ctx context.Context, qr sqlx.QueryerContext, f LeagueFilter) ([]League, error)

	InsertMatch(ctx context.Context, ec sqlx.ExecerContext, m Match) error
	SelectMatches(ctx context.Context, qr sqlx.QueryerContext, f MatchFilter, lock bool) ([]Match, error)
	UpdateMatch(ctx context.Context, ec sqlx.ExecerContext, m Match) error

	UpsertOrganizationAccount(ctx context.Context, ec sqlx.ExecerContext, oid, aid string) error
	UpsertAccount(ctx context.Context, ec sqlx.ExecerContext, a Account) error
	SelectAccounts(ctx context.Context, qr sqlx.QueryerContext, f AccountFilter) ([]Account, error)

	SelectMatchScouts(ctx context.Context, qr sqlx.QueryerContext, f MatchScoutFilter) ([]MatchScout, error)
	InsertMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms MatchScout) error
	UpdateMatchScout(ctx context.Context, ec sqlx.ExecerContext, ms MatchScout) error
}
