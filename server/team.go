package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/sportsbydata/backend/scouting"
)

type team struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

func newTeam(t scouting.Team) team {
	return team{
		UUID: t.UUID,
		Name: t.Name,
	}
}

func (rt *Server) createTeam(w http.ResponseWriter, r *http.Request) {
	var nt scouting.NewTeam

	if err := json.NewDecoder(r.Body).Decode(&nt); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	t, err := scouting.CreateTeam(r.Context(), nt, rt.sdb, rt.store)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newTeam(t))
}

func (rt *Server) getTeams(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		LeagueUUID uuid.UUID   `schema:"league_uuid"`
		TeamUUIDs  []uuid.UUID `schema:"team_uuids"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	f := scouting.TeamFilter{
		UUIDs:          qr.TeamUUIDs,
		OrganizationID: claims.ActiveOrganizationID,
		LeagueUUID:     qr.LeagueUUID,
	}

	tt, err := rt.store.SelectTeams(r.Context(), rt.sdb, f)
	if err != nil {
		HandleError(w, err)

		return
	}

	enc := make([]team, len(tt))

	for i, t := range tt {
		enc[i] = newTeam(t)
	}

	JSON(w, http.StatusOK, Paginated(enc, ""))
}
