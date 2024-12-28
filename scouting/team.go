package scouting

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
)

type NewTeam struct {
	Name string `json:"name"`
}

type TeamFilter struct {
	UUIDs          []uuid.UUID
	LeagueUUID     uuid.UUID
	OrganizationID string
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
