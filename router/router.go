package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
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

	group.HandleFunc("GET /account", rt.fetchAccount)

	return group
}

func (rt *Router) fetchAccount(w http.ResponseWriter, r *http.Request) {
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

	JSON(w, http.StatusCreated, encodeTeam(t))
}

func encodeTeam(t scouting.Team) any {
	return struct {
		UUID uuid.UUID `json:"uuid"`
		Name string    `json:"name"`
	}{
		UUID: t.UUID,
		Name: t.Name,
	}
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
