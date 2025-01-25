package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/sportsbydata/backend/scouting"
)

type league struct {
	UUID  uuid.UUID `json:"uuid"`
	Name  string    `json:"name"`
	Teams []team    `json:"teams,omitempty"`
}

func newLeague(l scouting.League, teams []scouting.Team) league {
	tt := make([]team, len(teams))

	for i, t := range teams {
		tt[i] = newTeam(t)
	}

	return league{
		UUID:  l.UUID,
		Name:  l.Name,
		Teams: tt,
	}
}

func (rt *Server) createLeague(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var nl scouting.NewLeague

	if err := json.NewDecoder(r.Body).Decode(&nl); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	l, err := scouting.CreateLeague(r.Context(), nl, rt.sdb)
	if err != nil {
		HandleError(w, err)

		return
	}

	tt, err := scouting.SelectTeams(r.Context(), rt.sdb, scouting.TeamFilter{
		LeagueUUID:     l.UUID,
		OrganizationID: claims.ActiveOrganizationID,
	})
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newLeague(l, tt))
}

func (rt *Server) updateOrganizationLeagues(w http.ResponseWriter, r *http.Request) {
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
		claims.ActiveOrganizationID,
		in.LeagueUUIDs,
	)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusOK, struct{}{})
}

func (rt *Server) getLeagues(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		LeagueUUID uuid.UUID `schema:"league_uuid"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	f := scouting.LeagueFilter{
		LeagueUUID:     qr.LeagueUUID,
		OrganizationID: claims.ActiveOrganizationID,
	}

	ll, err := scouting.SelectLeagues(r.Context(), rt.sdb, f)
	if err != nil {
		HandleError(w, err)

		return
	}

	leagueTeams := make(map[uuid.UUID][]scouting.Team)

	for _, l := range ll {
		tt, err := scouting.SelectTeams(r.Context(), rt.sdb, scouting.TeamFilter{
			OrganizationID: claims.ActiveOrganizationID,
			LeagueUUID:     l.UUID,
		})
		if err != nil {
			HandleError(w, err)

			return
		}

		leagueTeams[l.UUID] = tt
	}

	mapped := make([]league, len(ll))

	for i, l := range ll {
		mapped[i] = newLeague(l, leagueTeams[l.UUID])
	}

	JSON(w, http.StatusOK, mapped)
}
