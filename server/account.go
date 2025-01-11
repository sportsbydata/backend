package server

import (
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/sportsbydata/backend/scouting"
)

type account struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
}

func newAccount(a scouting.Account) account {
	return account{
		ID:        a.ID,
		FirstName: a.FirstName,
		LastName:  a.LastName,
		AvatarURL: a.AvatarURL,
	}
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	clerkUser, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		slog.Error("getting user", slog.Any("error", err))
		Internal(w)

		return
	}

	a, err := scouting.InsertAccount(r.Context(), s.sdb, s.store, claims.ActiveOrganizationID, clerkUser)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newAccount(a))
}

func (s *Server) getAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	f := scouting.AccountFilter{
		OrganizationID: claims.ActiveOrganizationID,
	}

	aa, err := s.store.SelectAccounts(r.Context(), s.sdb, f)
	if err != nil {
		HandleError(w, err)

		return
	}

	enc := make([]account, len(aa))

	for i, a := range aa {
		enc[i] = newAccount(a)
	}

	JSON(w, http.StatusOK, enc)
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	f := scouting.AccountFilter{
		OrganizationID: claims.ActiveOrganizationID,
		ID:             claims.Subject,
	}

	aa, err := s.store.SelectAccounts(r.Context(), s.sdb, f)
	if err != nil {
		HandleError(w, err)

		return
	}

	if len(aa) == 0 {
		NotFound(w)

		return
	}

	JSON(w, http.StatusOK, newAccount(aa[0]))
}
