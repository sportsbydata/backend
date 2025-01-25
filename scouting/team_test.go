package scouting

import (
	"context"

	"github.com/Masterminds/squirrel"
)

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
