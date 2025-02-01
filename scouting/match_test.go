package scouting

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/sportsbydata/backend/sbd"
	"github.com/stretchr/testify/assert"
)

func Test_modesSubmodesConflicts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		First     modeSubmode
		Second    modeSubmode
		Available []modeSubmode
	}{
		{
			First: modeSubmode{
				mode:    ModeAttack,
				submode: SubmodeOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeDefence,
				submode: SubmodeOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttack,
				submode: SubmodeNotOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeDefence,
				submode: SubmodeNotOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttackDefence,
				submode: SubmodeAnyRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttackDefence,
				submode: SubmodeAllRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		// TODO: add cases for plays
	}

	buildUnavailable := func(skip []modeSubmode) []modeSubmode {
		var mm []modeSubmode

		for _, m := range []Mode{ModeAttack, ModeDefence, ModeAttackDefence} {
			for _, sm := range []Submode{SubmodeOurRules, SubmodeNotOurRules, SubmodeAnyRules, SubmodeAllRules, SubmodePlays} {
				if !modeSubmodeValid(m, sm) {
					continue
				}

				var ignore bool

				for _, sk := range skip {
					if sk.mode == m && sk.submode == sm {
						ignore = true

						break
					}
				}

				if ignore {
					continue
				}

				mm = append(mm, modeSubmode{m, sm})
			}
		}

		return mm
	}

	for _, tc := range cases {
		baseName := fmt.Sprintf("(%s+%s)", tc.First.mode, tc.First.submode)

		for _, av := range tc.Available {
			fullName := fmt.Sprintf("%s with (%s+%s)", baseName, av.mode, av.submode)

			t.Run(fullName, func(t *testing.T) {
				assert.False(t, modesSubmodesConflicts(tc.First, av), "should not conflict")
			})
		}

		for _, sm := range buildUnavailable(tc.Available) {
			fullName := fmt.Sprintf("%s with (%s+%s)", baseName, sm.mode, sm.submode)

			t.Run(fullName, func(t *testing.T) {
				assert.True(t, modesSubmodesConflicts(tc.First, sm), "should conflict")
			})
		}
	}
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
		s.Assert().Equal(sbd.NewNotFoundError("league"), err)
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
		s.Assert().Equal(sbd.NewValidationError("team not found in league"), err)
	})
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

	a, err := OnboardAccount(context.Background(), s.sdb, "o1", clerkUser)
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

	a, err := OnboardAccount(context.Background(), s.sdb, "o1", clerkUser)
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
