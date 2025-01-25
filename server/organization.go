package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkorg "github.com/clerk/clerk-sdk-go/v2/organization"
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
	var in struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	var apierr *clerk.APIErrorResponse

	_, err := clerkorg.Get(r.Context(), in.ID)
	switch {
	case err == nil:
		// OK.
	case errors.As(err, &apierr):
		if apierr.Response.StatusCode == http.StatusNotFound {
			NotFound(w, "user not found in clerk")

			return
		}

		slog.Error("getting organization", slog.Any("error", err))
		Internal(w)

		return
	default:
		slog.Error("getting organization", slog.Any("error", err))
		Internal(w)

		return
	}

	o, err := scouting.CreateOrganization(r.Context(), s.sdb, in.ID)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newOrganization(o))
}

func (s *Server) getOrganization(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	filter := scouting.OrganizationFilter{
		IDs: []string{claims.ActiveOrganizationID},
	}

	oo, err := scouting.SelectOrganizations(r.Context(), s.sdb, filter)
	if err != nil {
		HandleError(w, err)

		return
	}

	if len(oo) == 0 {
		NotFound(w, "organization not found")

		return
	}

	JSON(w, http.StatusCreated, newOrganization(oo[0]))
}
