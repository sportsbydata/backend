package scouting

import (
	"context"

	"github.com/Masterminds/squirrel"
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
