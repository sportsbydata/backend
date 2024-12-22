package router

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
	UUID         uuid.UUID  `json:"uuid"`
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
	return match{
		UUID:         m.UUID,
		LeagueUUID:   m.LeagueUUID,
		AwayTeamUUID: m.AwayTeamUUID,
		HomeTeamUUID: m.HomeTeamUUID,
		CreatedBy:    m.CreatedBy,
		HomeScore:    m.HomeScore,
		AwayScore:    m.AwayScore,
		StartsAt:     m.StartsAt,
		FinishedAt:   m.FinishedAt,
	}
}

func (rt *Router) createMatch(w http.ResponseWriter, r *http.Request) {
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

	m, err := scouting.CreateMatch(r.Context(), rt.sdb, rt.db, claims.ActiveOrganizationID, claims.Subject, nm)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("creating match", slog.Any("error", err))

			return
		}

		return
	}

	JSON(w, http.StatusCreated, newMatch(m))
}

func (rt *Router) getMatches(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		Active bool `schema:"active"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	f := scouting.MatchFilter{
		Active:         &qr.Active,
		OrganizationID: &claims.ActiveOrganizationID,
	}

	mm, err := rt.db.SelectMatches(r.Context(), rt.sdb, f, false)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("selecting organization matches", slog.Any("error", err))
		}

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
	return matchScout{
		AccountID:  ms.AccountID,
		Mode:       ms.Mode,
		Submode:    ms.Submode,
		FinishedAt: ms.FinishedAt,
	}
}

func (rt *Router) getMatchScouts(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var qr struct {
		MatchUUID uuid.UUID `schema:"match_uuid"`
	}

	if err := rt.decoder.Decode(&qr, r.URL.Query()); err != nil {
		BadRequest(w, "invalid query")

		return
	}

	f := scouting.MatchScoutFilter{
		MatchUUID:           &qr.MatchUUID,
		MatchOrganizationID: &claims.ActiveOrganizationID,
	}

	mss, err := rt.db.SelectMatchScouts(r.Context(), rt.sdb, f)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("selecting match scouts", slog.Any("error", err))
		}

		return
	}

	enc := make([]matchScout, len(mss))

	for i, ms := range mss {
		enc[i] = newMatchScout(ms)
	}

	JSON(w, http.StatusOK, enc)
}
