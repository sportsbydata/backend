package router

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/go-pkgz/routegroup"
	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/db"
	"github.com/sportsbydata/backend/scouting"
)

type Router struct {
	sdb *sqlx.DB
	db  *db.DB
}

func New(sdb *sqlx.DB) *Router {
	return &Router{
		sdb: sdb,
		db:  &db.DB{},
	}
}

func (rt *Router) Handler() http.Handler {
	m := http.NewServeMux()

	group := routegroup.New(m)

	group.Use(clerkhttp.RequireHeaderAuthorization())

	group.HandleFunc("GET /me", rt.me)
	group.HandleFunc("GET /account", rt.getOrganizationAccounts)
	group.HandleFunc("POST /team", rt.createTeam)
	group.HandleFunc("POST /league", rt.createLeague)
	group.HandleFunc("GET /league", rt.getOrganizationLeagues)
	group.HandleFunc("GET /league/{league_uuid}/team", rt.getLeagueTeams)
	group.HandleFunc("PUT /organization-league", rt.updateOrganizationLeagues)
	group.HandleFunc("POST /match", rt.insertMatch)
	group.HandleFunc("GET /match", rt.getOrganizationMatches)
	group.HandleFunc("GET /match/{match_uuid}/match-scout", rt.getMatchScouts)

	return group
}

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

func (rt *Router) getOrganizationAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	oid := claims.ActiveOrganizationID
	if oid == "" {
		BadRequest(w, "organization not found")

		return
	}

	aa, err := scouting.SelectOrganizationAccounts(r.Context(), rt.sdb, &db.DB{}, oid)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("selecting organization accounts", slog.Any("error", err))
		}

		return
	}

	enc := make([]account, len(aa))

	for i, a := range aa {
		enc[i] = newAccount(a)
	}

	JSON(w, http.StatusOK, enc)
}

func (rt *Router) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	oid := claims.ActiveOrganizationID
	if oid == "" {
		BadRequest(w, "organization not found")

		return
	}

	u, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		slog.Error("getting user", slog.Any("error", err))
		Internal(w)

		return
	}

	tnow := time.Now()

	a := scouting.Account{
		ID:         claims.Subject,
		FirstName:  *u.FirstName,
		LastName:   *u.LastName,
		AvatarURL:  *u.ImageURL,
		CreatedAt:  tnow,
		ModifiedAt: tnow,
	}

	err = scouting.UpsertAccount(r.Context(), rt.sdb, &db.DB{}, oid, a)
	if err != nil {
		slog.Error("upserting account", slog.Any("error", err))

		return
	}

	JSON(w, http.StatusOK, a)
}

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

	l, err := scouting.CreateLeague(r.Context(), nl, rt.sdb, &db.DB{})
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

	if claims.ActiveOrganizationID == "" {
		BadRequest(w, "missing organization id")

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
		&db.DB{},
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

type match struct {
	UUID           uuid.UUID  `json:"uuid"`
	LeagueUUID     uuid.UUID  `json:"league_uuid"`
	AwayTeamUUID   uuid.UUID  `json:"away_team_uuid"`
	HomeTeamUUID   uuid.UUID  `json:"home_team_uuid"`
	CreatedBy      string     `json:"created_by"`
	HomeScore      *uint      `json:"home_score,omitempty"`
	AwayScore      *uint      `json:"away_score,omitempty"`
	OrganizationID string     `json:"organization_id"`
	StartsAt       time.Time  `json:"starts_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

func newMatch(m scouting.Match) match {
	return match{
		UUID:           m.UUID,
		LeagueUUID:     m.LeagueUUID,
		AwayTeamUUID:   m.AwayTeamUUID,
		HomeTeamUUID:   m.HomeTeamUUID,
		CreatedBy:      m.CreatedBy,
		HomeScore:      m.HomeScore,
		AwayScore:      m.AwayScore,
		OrganizationID: m.OrganizationID,
		StartsAt:       m.StartsAt,
		FinishedAt:     m.FinishedAt,
	}
}

func (rt *Router) insertMatch(w http.ResponseWriter, r *http.Request) {
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

	m, err := scouting.CreateMatch(r.Context(), rt.sdb, &db.DB{}, claims.ActiveOrganizationID, claims.Subject, nm)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("creating match", slog.Any("error", err))

			return
		}

		return
	}

	JSON(w, http.StatusCreated, newMatch(m))
}

func (rt *Router) getOrganizationLeagues(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	ll, err := scouting.SelectOrganizationLeagues(r.Context(), claims.ActiveOrganizationID, rt.sdb, &db.DB{})
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

func (rt *Router) getLeagueTeams(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	luuid, err := uuid.FromString(r.PathValue("league_uuid"))
	if err != nil {
		BadRequest(w, "invalid league uuid")

		return
	}

	l, err := scouting.GetOrganizationLeague(r.Context(), rt.sdb, &db.DB{}, claims.ActiveOrganizationID, luuid)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("getting organization league", slog.Any("error", err))
		}

		return
	}

	tt, err := scouting.SelectTeams(r.Context(), rt.sdb, &db.DB{}, scouting.TeamFilter{
		LeagueUUID: &l.UUID,
	})
	if err != nil {
		if CoreError(w, err) {
			slog.Error("selecting teams", slog.Any("error", err))
		}

		return
	}

	mapped := make([]team, len(tt))

	for i, t := range tt {
		mapped[i] = newTeam(t)
	}

	JSON(w, http.StatusOK, mapped)
}

func (rt *Router) getOrganizationMatches(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		slog.Error("session not found in context")
		Internal(w)

		return
	}

	var active bool

	if v := r.URL.Query().Get("active"); v != "" {
		bv, err := strconv.ParseBool(v)
		if err != nil {
			BadRequest(w, "invalid active")

			return
		}

		active = bv
	}

	mm, err := scouting.SelectOrganizationMatches(r.Context(), rt.sdb, &db.DB{}, claims.ActiveOrganizationID, scouting.MatchFilter{
		Active: active,
	})
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

	JSON(w, http.StatusOK, enc)
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

	muuid, err := uuid.FromString(r.PathValue("match_uuid"))
	if err != nil {
		BadRequest(w, "invalid match uuid")

		return
	}

	m, err := scouting.GetOrganizationMatch(r.Context(), rt.sdb, &db.DB{}, claims.ActiveOrganizationID, muuid)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("getting organization match", slog.Any("error", err))
		}

		return
	}

	mss, err := scouting.SelectMatchScouts(r.Context(), rt.sdb, &db.DB{}, m.UUID)
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

func (rt *Router) createTeam(w http.ResponseWriter, r *http.Request) {
	var nt scouting.NewTeam

	if err := json.NewDecoder(r.Body).Decode(&nt); err != nil {
		BadRequest(w, "invalid json")

		return
	}

	t, err := scouting.CreateTeam(r.Context(), nt, rt.sdb, rt.db)
	if err != nil {
		if CoreError(w, err) {
			slog.Error("creating team", slog.Any("error", err))
		}

		return
	}

	JSON(w, http.StatusCreated, newTeam(t))
}

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeError(w, "bad_request", http.StatusBadRequest, msg)
}

func Internal(w http.ResponseWriter) {
	writeError(w, "internal_error", http.StatusInternalServerError, "internal error")
}

func CoreError(w http.ResponseWriter, err error) (log bool) {
	var ve *scouting.ValidationError

	switch {
	case errors.As(err, &ve):
		BadRequest(w, err.Error())

		return
	}

	Internal(w)

	return true
}

func writeError(w http.ResponseWriter, code string, status int, msg string) {
	JSON(w, status, struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}{
		Message: msg,
		Code:    code,
	})
}
