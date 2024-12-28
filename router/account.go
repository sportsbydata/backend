package router

import (
	"log/slog"
	"net/http"
	"time"

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

func (rt *Router) getAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		Self bool `schema:"self"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	var aa []scouting.Account

	if qr.Self {
		var (
			ok bool
			a  scouting.Account
		)

		if a, ok = rt.me(w, r); !ok {
			return
		}

		aa = []scouting.Account{a}
	} else {
		f := scouting.AccountFilter{
			OrganizationID: &claims.ActiveOrganizationID,
		}

		var err error

		if aa, err = rt.db.SelectAccounts(r.Context(), rt.sdb, f); err != nil {
			HandleError(w, err)

			return
		}
	}

	enc := make([]account, len(aa))

	for i, a := range aa {
		enc[i] = newAccount(a)
	}

	JSON(w, http.StatusOK, Paginated(enc, ""))
}

func (rt *Router) me(w http.ResponseWriter, r *http.Request) (scouting.Account, bool) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return scouting.Account{}, false
	}

	u, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		slog.Error("getting user", slog.Any("error", err))
		Internal(w)

		return scouting.Account{}, false
	}

	tnow := time.Now()

	logger := slog.With(slog.String("account_id", claims.Subject))

	if u.FirstName == nil {
		logger.Error("attempted to upsert account with nil first name")
		Internal(w)

		return scouting.Account{}, false
	}

	if u.LastName == nil {
		logger.Error("attempted to upsert account with nil last name")
		Internal(w)

		return scouting.Account{}, false
	}

	if u.ImageURL == nil {
		logger.Error("attempted to upsert account with nil image url")
		Internal(w)

		return scouting.Account{}, false
	}

	a := scouting.Account{
		ID:         claims.Subject,
		FirstName:  *u.FirstName,
		LastName:   *u.LastName,
		AvatarURL:  *u.ImageURL,
		CreatedAt:  tnow,
		ModifiedAt: tnow,
	}

	err = scouting.UpsertAccount(r.Context(), rt.sdb, rt.db, claims.ActiveOrganizationID, a)
	if err != nil {
		slog.Error("upserting account", slog.Any("error", err))

		return scouting.Account{}, false
	}

	return a, true
}
