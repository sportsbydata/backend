package scouting

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
)

func (s *Suite) Test_CreateLeague() {
	t1, err := CreateTeam(context.Background(), NewTeam{
		Name: "t1",
	}, s.sdb)
	s.Require().NoError(err)

	t2, err := CreateTeam(context.Background(), NewTeam{
		Name: "t2",
	}, s.sdb)
	s.Require().NoError(err)

	l, err := CreateLeague(context.Background(), NewLeague{
		Name: "league",
		TeamUUIDs: []uuid.UUID{
			t1.UUID,
			t2.UUID,
		},
	}, s.sdb)
	s.Require().NoError(err)

	s.Assert().NotEmpty(l.UUID)
	s.Assert().NotEmpty(l.CreatedAt)
	s.Assert().NotEmpty(l.ModifiedAt)

	cnt := s.selectCount("league", squirrel.Eq{
		"uuid": l.UUID,
	})
	s.Assert().Equal(1, cnt)

	cnt = s.selectCount("league_team", squirrel.Eq{
		"league_uuid": l.UUID,
	})

	s.Assert().Equal(2, cnt)
}

func (s *Suite) Test_UpdateOrganizationLeagues() {
	t1, err := CreateTeam(context.Background(), NewTeam{
		Name: "t1",
	}, s.sdb)
	s.Require().NoError(err)

	t2, err := CreateTeam(context.Background(), NewTeam{
		Name: "t2",
	}, s.sdb)
	s.Require().NoError(err)

	l1, err := CreateLeague(context.Background(), NewLeague{
		Name: "league",
		TeamUUIDs: []uuid.UUID{
			t1.UUID,
		},
	}, s.sdb)
	s.Require().NoError(err)

	l2, err := CreateLeague(context.Background(), NewLeague{
		Name: "league",
		TeamUUIDs: []uuid.UUID{
			t2.UUID,
		},
	}, s.sdb)
	s.Require().NoError(err)

	_, err = CreateOrganization(context.Background(), s.sdb, "o1")
	s.Require().NoError(err)

	err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l1.UUID, l2.UUID})
	s.Require().NoError(err)

	cnt := s.selectCount("organization_league", squirrel.Eq{"organization_id": "o1"})
	s.Assert().Equal(2, cnt)

	err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l1.UUID})
	s.Require().NoError(err)

	cnt = s.selectCount("organization_league", squirrel.Eq{"organization_id": "o1"})
	s.Assert().Equal(1, cnt)
}
