package router

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-pkgz/routegroup"
	"github.com/gofrs/uuid/v5"
	"github.com/iris-contrib/schema"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/db"
	"github.com/sportsbydata/backend/scouting"
)

var decoder = schema.NewDecoder()

type Router struct {
	sdb     *sqlx.DB
	db      *db.DB
	decoder *schema.Decoder
}

func New(sdb *sqlx.DB) *Router {
	dec := schema.NewDecoder()

	dec.RegisterConverter(uuid.UUID{}, func(s string) reflect.Value {
		u, err := uuid.FromString(s)
		if err != nil {
			return reflect.Value{}
		}

		return reflect.ValueOf(u)
	})

	return &Router{
		sdb:     sdb,
		db:      &db.DB{},
		decoder: dec,
	}
}

func (rt *Router) Handler() http.Handler {
	m := http.NewServeMux()

	group := routegroup.New(m)

	group.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("header", slog.String("h", r.Header.Get("Authorization")))

			h.ServeHTTP(w, r)
		})
	})
	group.Use(withOrg)

	group.HandleFunc("GET /account", rt.getAccounts)
	group.HandleFunc("GET /league", rt.getLeagues)
	group.HandleFunc("POST /league", rt.createLeague)
	group.HandleFunc("POST /team", rt.createTeam)
	group.HandleFunc("GET /team", rt.getTeams)
	group.HandleFunc("PUT /organization-league", rt.updateOrganizationLeagues)
	group.HandleFunc("POST /match", rt.createMatch)
	group.HandleFunc("GET /match", rt.getMatches)
	group.HandleFunc("GET /match-scout", rt.getMatchScouts)
	group.HandleFunc("POST /match-scout", rt.createMatchScout)
	group.HandleFunc("PATCH /match-scout", rt.updateMatchScout)

	return group
}

func Paginated[T any](items []T, cursor string) any {
	return struct {
		Items  []T    `json:"items"`
		Cursor string `json:"cursor"`
	}{
		Items:  items,
		Cursor: cursor,
	}
}

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}

func Unauthorized(w http.ResponseWriter) {
	writeError(w, "unauthorized", http.StatusUnauthorized, "unauthorized")
}

func NotFound(w http.ResponseWriter) {
	writeError(w, "not_found", http.StatusBadRequest, "not_found")
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeError(w, "bad_request", http.StatusBadRequest, msg)
}

func Internal(w http.ResponseWriter) {
	writeError(w, "internal_error", http.StatusInternalServerError, "internal error")
}

func HandleError(w http.ResponseWriter, err error) {
	var ve *scouting.ValidationError

	switch {
	case errors.As(err, &ve):
		BadRequest(w, err.Error())

		return
	case errors.Is(err, scouting.ErrStoreNotFound):
		NotFound(w)

		return
	}

	Internal(w)
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
