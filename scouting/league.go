package scouting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

type NewLeague struct {
	Name      string      `json:"name"`
	TeamUUIDs []uuid.UUID `json:"team_uuids"`
}

func (nl *NewLeague) ToLeague() League {
	tnow := time.Now()

	return League{
		UUID:       uuid.Must(uuid.NewV7()),
		Name:       nl.Name,
		CreatedAt:  tnow,
		ModifiedAt: tnow,
	}
}

type League struct {
	UUID uuid.UUID `db:"league.uuid"`
	Name string    `db:"league.name"`

	CreatedAt  time.Time `db:"league.created_at"`
	ModifiedAt time.Time `db:"league.modified_at"`
}

type NewTeam struct {
	Name string `json:"name"`
}

func (nt *NewTeam) ToTeam() Team {
	tnow := time.Now()

	return Team{
		UUID:       uuid.Must(uuid.NewV7()),
		Name:       nt.Name,
		CreatedAt:  tnow,
		ModifiedAt: tnow,
	}
}

type Team struct {
	UUID uuid.UUID `db:"team.uuid"`
	Name string    `db:"team.name"`

	CreatedAt  time.Time `db:"team.created_at"`
	ModifiedAt time.Time `db:"team.modified_at"`
}

func CreateTeam(ctx context.Context, nt NewTeam, sdb *sqlx.DB, store Store) (Team, error) {
	t := nt.ToTeam()

	if err := store.InsertTeam(ctx, sdb, t); err != nil {
		return Team{}, fmt.Errorf("inserting team: %w", err)
	}

	return t, nil
}

func UpdateOrganizationLeagues(ctx context.Context, sdb *sqlx.DB, store Store, oid string, luuids []uuid.UUID) error {
	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = store.DeleteOrganizationLeagues(ctx, tx, oid); err != nil {
		return err
	}

	for _, lu := range luuids {
		if err = store.InsertOrganizationLeague(ctx, tx, oid, lu); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func CreateLeague(ctx context.Context, nl NewLeague, sdb *sqlx.DB, store Store) (League, error) {
	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		return League{}, err
	}

	defer tx.Rollback()

	tt, err := store.SelectTeams(ctx, tx, TeamFilter{
		UUIDs: nl.TeamUUIDs,
	})
	if err != nil {
		return League{}, err
	}

	if len(tt) != len(nl.TeamUUIDs) {
		return League{}, fmt.Errorf("expected %d teams, found %d", len(nl.TeamUUIDs), len(tt))
	}

	l := nl.ToLeague()

	if err := store.InsertLeague(ctx, tx, l); err != nil {
		return League{}, fmt.Errorf("inserting league: %w", err)
	}

	for _, t := range tt {
		if err := store.InsertLeagueTeam(ctx, tx, l.UUID, t.UUID); err != nil {
			return League{}, fmt.Errorf("inserting league team: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return League{}, fmt.Errorf("commiting: %w", err)
	}

	return l, nil
}

type TeamFilter struct {
	UUIDs          []uuid.UUID
	LeagueUUID     *uuid.UUID
	OrganizationID *string
}

type LeagueFilter struct {
	LeagueUUID     *uuid.UUID
	OrganizationID *string
}

var (
	ErrStoreNotFound = errors.New("not found")
)

type MatchScoutFilter struct {
	MatchUUID           *uuid.UUID
	MatchOrganizationID *string
}

type Store interface {
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
}
