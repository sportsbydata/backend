package scouting

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

type NewLeague struct {
	Name      string      `json:"name"`
	TeamUUIDs []uuid.UUID `json:"team_uuids"`
}

type LeagueFilter struct {
	LeagueUUID     uuid.UUID
	OrganizationID string
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

func UpdateOrganizationLeagues(ctx context.Context, sdb *sqlx.DB, oid string, luuids []uuid.UUID) error {
	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = deleteOrganizationLeagues(ctx, tx, oid); err != nil {
		return err
	}

	for _, lu := range luuids {
		if err = insertOrganizationLeague(ctx, tx, oid, lu); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func CreateLeague(ctx context.Context, nl NewLeague, sdb *sqlx.DB) (League, error) {
	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		return League{}, err
	}

	defer tx.Rollback()

	tt, err := SelectTeams(ctx, tx, TeamFilter{
		UUIDs: nl.TeamUUIDs,
	})
	if err != nil {
		return League{}, err
	}

	if len(tt) != len(nl.TeamUUIDs) {
		return League{}, fmt.Errorf("expected %d teams, found %d", len(nl.TeamUUIDs), len(tt))
	}

	l := nl.ToLeague()

	if err := insertLeague(ctx, tx, l); err != nil {
		return League{}, fmt.Errorf("inserting league: %w", err)
	}

	for _, t := range tt {
		if err := insertLeagueTeam(ctx, tx, l.UUID, t.UUID); err != nil {
			return League{}, fmt.Errorf("inserting league team: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return League{}, fmt.Errorf("commiting: %w", err)
	}

	return l, nil
}
