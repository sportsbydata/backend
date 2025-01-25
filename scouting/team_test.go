package scouting

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
)

func (s *Suite) Test_CreateOrganization() {
	o, err := CreateOrganization(context.Background(), s.sdb, "id")
	s.Require().NoError(err)

	s.Assert().Equal("id", o.ID)
	s.Assert().Equal(DefaultScoutingConfig, o.ScoutingConfig)
	s.Assert().NotZero(o.CreatedAt)
	s.Assert().NotZero(o.ModifiedAt)

	cnt := s.selectCount("organization", squirrel.Eq{
		"id": o.ID,
	})

	s.Assert().Equal(1, cnt)
}

func (s *Suite) Test_CreateTeam() {
	t, err := CreateTeam(context.Background(), NewTeam{
		Name: "test",
	}, s.sdb)
	s.Require().NoError(err)

	s.Assert().Equal("test", t.Name)
	s.Assert().NotEmpty(t.UUID)
	s.Assert().NotEmpty(t.ModifiedAt)
	s.Assert().NotEmpty(t.CreatedAt)

	cnt := s.selectCount("team", squirrel.Eq{
		"uuid": t.UUID,
	})

	s.Assert().Equal(1, cnt)
}

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

func (s *Suite) Test_CreateMatch() {
	s.Run("success", func() {
		s.TearDownTest()

		home, err := CreateTeam(context.Background(), NewTeam{
			Name: "home",
		}, s.sdb)
		s.Require().NoError(err)

		away, err := CreateTeam(context.Background(), NewTeam{
			Name: "away",
		}, s.sdb)
		s.Require().NoError(err)

		l, err := CreateLeague(context.Background(), NewLeague{
			Name: "league",
			TeamUUIDs: []uuid.UUID{
				home.UUID,
				away.UUID,
			},
		}, s.sdb)
		s.Require().NoError(err)

		_, err = CreateOrganization(context.Background(), s.sdb, "o1")
		s.Require().NoError(err)

		err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l.UUID})
		s.Require().NoError(err)

		starts := time.Now().Add(time.Hour)

		nm := NewMatch{
			LeagueUUID:   l.UUID,
			AwayTeamUUID: away.UUID,
			HomeTeamUUID: home.UUID,
			StartsAt:     starts,
		}

		m, err := CreateMatch(context.Background(), s.sdb, "o1", "test_scout", nm)
		s.Require().NoError(err)

		s.Assert().NotEmpty(m.UUID)
		s.Assert().NotEmpty(m.CreatedBy)
		s.Assert().NotEmpty(m.ModifiedAt)
		s.Assert().Equal(home.UUID, m.HomeTeamUUID)
		s.Assert().Equal(away.UUID, m.AwayTeamUUID)
		s.Assert().Equal(l.UUID, m.LeagueUUID)
		s.Assert().Equal("test_scout", m.CreatedBy)
		s.Assert().Equal(starts, m.StartsAt)

		cnt := s.selectCount("match", squirrel.Eq{"uuid": m.UUID})
		s.Assert().Equal(1, cnt)
	})

	s.Run("create match with league that is not linked with organization", func() {
		s.TearDownTest()

		home, err := CreateTeam(context.Background(), NewTeam{
			Name: "home",
		}, s.sdb)
		s.Require().NoError(err)

		away, err := CreateTeam(context.Background(), NewTeam{
			Name: "away",
		}, s.sdb)
		s.Require().NoError(err)

		l, err := CreateLeague(context.Background(), NewLeague{
			Name: "league",
			TeamUUIDs: []uuid.UUID{
				home.UUID,
				away.UUID,
			},
		}, s.sdb)
		s.Require().NoError(err)

		_, err = CreateOrganization(context.Background(), s.sdb, "o2")
		s.Require().NoError(err)

		err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o2", []uuid.UUID{l.UUID})
		s.Require().NoError(err)

		starts := time.Now().Add(time.Hour)

		nm := NewMatch{
			LeagueUUID:   l.UUID,
			AwayTeamUUID: away.UUID,
			HomeTeamUUID: home.UUID,
			StartsAt:     starts,
		}

		_, err = CreateMatch(context.Background(), s.sdb, "o1", "test_scout", nm)
		s.Assert().Equal(NewNotFoundError("league not found"), err)
	})

	s.Run("creating match with a team that does not belong to the league", func() {
		s.TearDownTest()

		home, err := CreateTeam(context.Background(), NewTeam{
			Name: "home",
		}, s.sdb)
		s.Require().NoError(err)

		away, err := CreateTeam(context.Background(), NewTeam{
			Name: "away",
		}, s.sdb)
		s.Require().NoError(err)

		l, err := CreateLeague(context.Background(), NewLeague{
			Name: "league",
			TeamUUIDs: []uuid.UUID{
				home.UUID,
			},
		}, s.sdb)
		s.Require().NoError(err)

		_, err = CreateOrganization(context.Background(), s.sdb, "o1")
		s.Require().NoError(err)

		err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l.UUID})
		s.Require().NoError(err)

		starts := time.Now().Add(time.Hour)

		nm := NewMatch{
			LeagueUUID:   l.UUID,
			AwayTeamUUID: away.UUID,
			HomeTeamUUID: home.UUID,
			StartsAt:     starts,
		}

		_, err = CreateMatch(context.Background(), s.sdb, "o1", "test_scout", nm)
		s.Assert().Equal(NewValidationError("team not found in league"), err)
	})
}

func (s *Suite) Test_InsertUser() {
	name := "matas"
	lastName := "ram"
	avatarURL := "https://google.com"

	clerkUser := &clerk.User{
		ID:        "1",
		FirstName: &name,
		LastName:  &lastName,
		ImageURL:  &avatarURL,
	}

	_, err := CreateOrganization(context.Background(), s.sdb, "o1")
	s.Require().NoError(err)

	_, err = InsertAccount(context.Background(), s.sdb, "o1", clerkUser)
	s.Require().NoError(err)

	cnt := s.selectCount("account", squirrel.Eq{"id": clerkUser.ID})
	s.Assert().Equal(1, cnt)

	cnt = s.selectCount("organization_account", squirrel.And{
		squirrel.Eq{"account_id": clerkUser.ID},
		squirrel.Eq{"organization_id": "o1"},
	})
	s.Assert().Equal(1, cnt)

	name = "john"
	lastName = "mayor"
	avatarURL = "https://x.com"

	clerkUser = &clerk.User{
		ID:        "1",
		FirstName: &name,
		LastName:  &lastName,
		ImageURL:  &avatarURL,
	}

	_, err = InsertAccount(context.Background(), s.sdb, "o1", clerkUser)
	s.Assert().Equal(ErrAlreadyExists, err)
}

func (s *Suite) Test_ScoutMatch() {
	name := "john"
	lastName := "mayor"
	avatarURL := "https://x.com"

	clerkUser := &clerk.User{
		ID:        "1",
		FirstName: &name,
		LastName:  &lastName,
		ImageURL:  &avatarURL,
	}

	_, err := CreateOrganization(context.Background(), s.sdb, "o1")
	s.Require().NoError(err)

	a, err := InsertAccount(context.Background(), s.sdb, "o1", clerkUser)
	s.Require().NoError(err)

	home, err := CreateTeam(context.Background(), NewTeam{
		Name: "home",
	}, s.sdb)
	s.Require().NoError(err)

	away, err := CreateTeam(context.Background(), NewTeam{
		Name: "away",
	}, s.sdb)
	s.Require().NoError(err)

	l, err := CreateLeague(context.Background(), NewLeague{
		Name: "league",
		TeamUUIDs: []uuid.UUID{
			home.UUID,
			away.UUID,
		},
	}, s.sdb)
	s.Require().NoError(err)

	err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l.UUID})
	s.Require().NoError(err)

	starts := time.Now().Add(time.Hour)

	nm := NewMatch{
		LeagueUUID:   l.UUID,
		AwayTeamUUID: away.UUID,
		HomeTeamUUID: home.UUID,
		StartsAt:     starts,
	}

	m, err := CreateMatch(context.Background(), s.sdb, "o1", "test_scout", nm)
	s.Require().NoError(err)

	err = ScoutMatch(context.Background(), s.sdb, "o1", a.ID, m.UUID, NewMatchScout{
		Mode:    ModeAttack,
		Submode: SubmodeAllRules,
	})
	s.Require().NoError(err)

	cnt := s.selectCount("match_scout", squirrel.And{
		squirrel.Eq{"account_id": a.ID},
		squirrel.Eq{"match_uuid": m.UUID},
		squirrel.Eq{"mode": ModeAttack},
		squirrel.Eq{"submode": SubmodeAllRules},
	})
	s.Assert().Equal(1, cnt)
}

func (s *Suite) Test_FinishMatch() {
	name := "john"
	lastName := "mayor"
	avatarURL := "https://x.com"

	clerkUser := &clerk.User{
		ID:        "1",
		FirstName: &name,
		LastName:  &lastName,
		ImageURL:  &avatarURL,
	}

	_, err := CreateOrganization(context.Background(), s.sdb, "o1")
	s.Require().NoError(err)

	a, err := InsertAccount(context.Background(), s.sdb, "o1", clerkUser)
	s.Require().NoError(err)

	home, err := CreateTeam(context.Background(), NewTeam{
		Name: "home",
	}, s.sdb)
	s.Require().NoError(err)

	away, err := CreateTeam(context.Background(), NewTeam{
		Name: "away",
	}, s.sdb)
	s.Require().NoError(err)

	l, err := CreateLeague(context.Background(), NewLeague{
		Name: "league",
		TeamUUIDs: []uuid.UUID{
			home.UUID,
			away.UUID,
		},
	}, s.sdb)
	s.Require().NoError(err)

	err = UpdateOrganizationLeagues(context.Background(), s.sdb, "o1", []uuid.UUID{l.UUID})
	s.Require().NoError(err)

	starts := time.Now().Add(time.Hour)

	nm := NewMatch{
		LeagueUUID:   l.UUID,
		AwayTeamUUID: away.UUID,
		HomeTeamUUID: home.UUID,
		StartsAt:     starts,
	}

	m, err := CreateMatch(context.Background(), s.sdb, "o1", "test_scout", nm)
	s.Require().NoError(err)

	err = ScoutMatch(context.Background(), s.sdb, "o1", a.ID, m.UUID, NewMatchScout{
		Mode:    ModeAttack,
		Submode: SubmodeAllRules,
	})
	s.Require().NoError(err)

	_, err = SubmitScoutReport(context.Background(), s.sdb, "o1", a.ID, m.UUID, ScoutReport{})
	s.Require().NoError(err)

	m, err = FinishMatch(context.Background(), s.sdb, "o1", m.UUID, MatchFinishRequest{
		HomeScore: 20,
		AwayScore: 30,
	})
	s.Require().NoError(err)

	s.Assert().Equal(uint(20), m.HomeScore.V)
	s.Assert().Equal(uint(30), m.AwayScore.V)
	s.Assert().NotEmpty(m.ModifiedAt)
	s.Assert().NotNil(m.FinishedAt)

	cnt := s.selectCount("match", squirrel.And{
		squirrel.Eq{"uuid": m.UUID},
		squirrel.Eq{"home_score": 20},
		squirrel.Eq{"away_score": 30},
	})
	s.Assert().Equal(1, cnt)
}
