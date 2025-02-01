package scouting

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/sportsbydata/backend/sbd"
)

func (s *Suite) Test_InsertAccount() {
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

	_, err = OnboardAccount(context.Background(), s.sdb, "o1", clerkUser)
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

	_, err = OnboardAccount(context.Background(), s.sdb, "o1", clerkUser)
	s.Assert().Equal(sbd.ErrAlreadyExists, err)
}
