package server

import (
	"net/http"

	"github.com/sportsbydata/backend/scouting"
)

type organization struct {
	ID             string         `json:"id"`
	ScoutingConfig scoutingConfig `json:"scouting_config"`
}

func newOrganization(o scouting.Organization) organization {
	return organization{
		ID:             o.ID,
		ScoutingConfig: newScoutingConfig(o.ScoutingConfig),
	}
}

func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "missing id")

		return
	}

	o, err := scouting.CreateOrganization(r.Context(), s.sdb, s.store, id)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newOrganization(o))
}
