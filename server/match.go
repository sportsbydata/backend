package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/sportsbydata/backend/scouting"
)

type match struct {
	ID           string     `json:"id"`
	LeagueUUID   uuid.UUID  `json:"league_uuid"`
	AwayTeamUUID uuid.UUID  `json:"away_team_uuid"`
	HomeTeamUUID uuid.UUID  `json:"home_team_uuid"`
	CreatedBy    string     `json:"created_by"`
	HomeScore    *uint      `json:"home_score,omitempty"`
	AwayScore    *uint      `json:"away_score,omitempty"`
	StartsAt     time.Time  `json:"starts_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

func newMatch(m scouting.Match) match {
	enc := match{
		ID:           m.UUID.String(),
		LeagueUUID:   m.LeagueUUID,
		AwayTeamUUID: m.AwayTeamUUID,
		HomeTeamUUID: m.HomeTeamUUID,
		CreatedBy:    m.CreatedBy,
		StartsAt:     m.StartsAt,
	}

	if m.HomeScore.Valid {
		enc.HomeScore = &m.HomeScore.V
	}

	if m.AwayScore.Valid {
		enc.AwayScore = &m.AwayScore.V
	}

	if m.FinishedAt.Valid {
		enc.FinishedAt = &m.FinishedAt.V
	}

	return enc
}

func (rt *Server) createMatch(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var nm scouting.NewMatch

	if err := json.NewDecoder(r.Body).Decode(&nm); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	m, err := scouting.CreateMatch(r.Context(), rt.sdb, rt.store, claims.ActiveOrganizationID, claims.Subject, nm)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusCreated, newMatch(m))
}

func (rt *Server) editMatch(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	matchUUID, err := uuid.FromString(r.PathValue("matchID"))
	if err != nil {
		BadRequest(w, "invalid match identifier format")

		return
	}

	var fr scouting.MatchFinishRequest

	if err := json.NewDecoder(r.Body).Decode(&fr); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	m, err := scouting.FinishMatch(
		r.Context(),
		rt.sdb,
		rt.store,
		claims.ActiveOrganizationID,
		matchUUID,
		fr,
	)
	if err != nil {
		HandleError(w, err)

		return
	}

	JSON(w, http.StatusOK, newMatch(m))
}

func (rt *Server) getMatches(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	f := scouting.MatchFilter{
		OrganizationID: claims.ActiveOrganizationID,
	}

	mm, err := rt.store.SelectMatches(r.Context(), rt.sdb, f, false)
	if err != nil {
		HandleError(w, err)

		return
	}

	enc := make([]match, len(mm))

	for i, m := range mm {
		enc[i] = newMatch(m)
	}

	JSON(w, http.StatusOK, Paginated(enc, ""))
}

type matchScout struct {
	AccountID  string           `json:"account_id"`
	Mode       scouting.Mode    `json:"mode"`
	Submode    scouting.Submode `json:"submode"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
}

func newMatchScout(ms scouting.MatchScout) matchScout {
	enc := matchScout{
		AccountID: ms.AccountID,
		Mode:      ms.Mode,
		Submode:   ms.Submode,
	}

	if ms.FinishedAt.Valid {
		enc.FinishedAt = &ms.FinishedAt.V
	}

	return enc
}

func (rt *Server) getMatchScouts(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	matchUUID, err := uuid.FromString(r.PathValue("matchID"))
	if err != nil {
		BadRequest(w, "invalid match identifier format")

		return
	}

	f := scouting.MatchScoutFilter{
		MatchUUID:           &matchUUID,
		MatchOrganizationID: &claims.ActiveOrganizationID,
	}

	mss, err := rt.store.SelectMatchScouts(r.Context(), rt.sdb, f)
	if err != nil {
		HandleError(w, err)

		return
	}

	enc := make([]matchScout, len(mss))

	for i, ms := range mss {
		enc[i] = newMatchScout(ms)
	}

	JSON(w, http.StatusOK, enc)
}

func (rt *Server) createMatchScout(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	matchUUID, err := uuid.FromString(r.PathValue("matchID"))
	if err != nil {
		BadRequest(w, "invalid match identifier format")

		return
	}

	var req scouting.NewMatchScout

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	if err := scouting.ScoutMatch(r.Context(), rt.sdb, rt.store, claims.ActiveOrganizationID, claims.Subject, matchUUID, req); err != nil {
		HandleError(w, err)

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (rt *Server) updateMatchScout(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	matchUUID, err := uuid.FromString(r.PathValue("matchID"))
	if err != nil {
		BadRequest(w, "invalid match identifier format")

		return
	}

	var req struct {
		Finished *bool `json:"finished"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	if req.Finished != nil && *req.Finished {
		ms, err := scouting.SubmitScoutReport(
			r.Context(),
			rt.sdb,
			rt.store,
			claims.ActiveOrganizationID,
			claims.Subject,
			matchUUID,
			scouting.ScoutReport{},
		)
		if err != nil {
			HandleError(w, err)

			return
		}

		JSON(w, http.StatusOK, newMatchScout(ms))

		return
	}

	w.WriteHeader(http.StatusOK)
}
