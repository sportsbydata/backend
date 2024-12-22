package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/sportsbydata/backend/scouting"
)

type league struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

func newLeague(l scouting.League) league {
	return league{
		UUID: l.UUID,
		Name: l.Name,
	}
}

func (rt *Router) createLeague(w http.ResponseWriter, r *http.Request) {
	var nl scouting.NewLeague

	if err := json.NewDecoder(r.Body).Decode(&nl); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	l, err := scouting.CreateLeague(r.Context(), nl, rt.sdb, rt.db)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("creating league", slog.Any("error", err))
		}

		return
	}

	JSON(w, http.StatusCreated, newLeague(l))
}

func (rt *Router) updateOrganizationLeagues(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var in struct {
		LeagueUUIDs []uuid.UUID `json:"league_uuids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	err := scouting.UpdateOrganizationLeagues(
		r.Context(),
		rt.sdb,
		rt.db,
		claims.ActiveOrganizationID,
		in.LeagueUUIDs,
	)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("updating organizaton leagues", slog.Any("error", err))
		}

		return
	}

	JSON(w, http.StatusOK, struct{}{})
}

func (rt *Router) getLeagues(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		LeagueUUID *uuid.UUID `schema:"league_uuid"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	f := scouting.LeagueFilter{
		LeagueUUID:     qr.LeagueUUID,
		OrganizationID: &claims.ActiveOrganizationID,
	}

	ll, err := rt.db.SelectLeagues(r.Context(), rt.sdb, f)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("selecting organization leagues", slog.Any("error", err))
		}

		return
	}

	mapped := make([]league, len(ll))

	for i, l := range ll {
		mapped[i] = newLeague(l)
	}

	JSON(w, http.StatusOK, mapped)
}
