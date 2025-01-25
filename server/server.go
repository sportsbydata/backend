package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-pkgz/routegroup"
	"github.com/gofrs/uuid/v5"
	"github.com/iris-contrib/schema"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sportsbydata/backend/access"
	"github.com/sportsbydata/backend/scouting"
)

//go:embed static/*
var static embed.FS

type Server struct {
	sdb     *sqlx.DB
	decoder *schema.Decoder
	hserver *http.Server
	dev     bool
	promKey string

	wg      sync.WaitGroup
	closeCh chan struct{}
}

func New(sdb *sqlx.DB, addr string, promKey string, dev bool) *Server {
	dec := schema.NewDecoder()

	dec.RegisterConverter(uuid.UUID{}, func(s string) reflect.Value {
		u, err := uuid.FromString(s)
		if err != nil {
			return reflect.Value{}
		}

		return reflect.ValueOf(u)
	})

	s := &Server{
		sdb:     sdb,
		decoder: dec,
		closeCh: make(chan struct{}),
		promKey: promKey,
		dev:     dev,
	}

	s.hserver = &http.Server{
		Addr:    addr,
		Handler: s.handler(),
	}

	return s
}

func (s *Server) Run() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		slog.Info("starting server", slog.Any("addr", s.hserver.Addr))

		err := s.hserver.ListenAndServe()
		switch {
		case err == nil, errors.Is(err, http.ErrServerClosed):
			// OK.
		default:
			slog.Error("listening server", slog.Any("error", err))
		}
	}()
}

func (s *Server) Close(ctx context.Context) error {
	slog.Info("shutting down server")

	if err := s.hserver.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("server shut down")

	s.wg.Wait()

	return nil
}

func (rt *Server) handler() http.Handler {
	m := http.NewServeMux()

	group := routegroup.New(m)

	if rt.dev {
		subbed, err := fs.Sub(static, "static")
		if err != nil {
			slog.Error("creating static sub", slog.Any("error", err))
		} else {
			fileSrv := http.FileServer(http.FS(subbed))

			group.Handle("/static/", http.StripPrefix("/static", fileSrv))
		}
	}

	if rt.promKey != "" {
		hdl := promhttp.Handler()

		group.Handle("/metrics", withApiKey(rt.promKey)(hdl))

		slog.Info("exposing metrics on /metrics")
	}

	group.Mount("/v1").Route(func(b *routegroup.Bundle) {
		b.With(withOrg).HandleFunc("POST /organizations", rt.createOrganization)
		b.With(withOrg).HandleFunc("GET /organization", rt.getOrganization)

		b.With(withOrg).HandleFunc("POST /accounts", rt.createAccount)
		b.With(withOrg).HandleFunc("GET /account", rt.getAccount)
		b.With(withOrg).HandleFunc("GET /accounts", rt.getAccounts)

		b.With(withOrgPerm(access.PermissionManageTeams)).HandleFunc("POST /teams", rt.createTeam)
		b.With(withOrg).HandleFunc("GET /teams", rt.getTeams)

		b.With(withOrg).HandleFunc("GET /leagues", rt.getLeagues)
		b.With(withOrgPerm(access.PermissionManageLeagues)).HandleFunc("POST /leagues", rt.createLeague)
		b.With(withOrgPerm(access.PermissionManageLeagues)).HandleFunc("PUT /organization/leagues", rt.updateOrganizationLeagues)

		b.HandleFunc("POST /matches", rt.createMatch)
		b.With(withOrg).HandleFunc("GET /matches/finished", rt.getFinishedMatches)
		b.With(withOrg).HandleFunc("GET /matches/active", rt.getActiveMatches)
		b.HandleFunc("POST /matches/{matchID}/finish", rt.finishMatch)
		b.HandleFunc("GET /matches/{matchID}", rt.getMatch)

		b.With(withOrg).HandleFunc("GET /matches/{matchID}/scouts", rt.getMatchScouts)
		b.HandleFunc("POST /matches/{matchID}/scout", rt.scoutMatch)
		b.HandleFunc("POST /matches/{matchID}/finish-scouting", rt.finishMatchScouting)
	})

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

func Forbidden(w http.ResponseWriter) {
	writeError(w, "forbidden", http.StatusForbidden, "forbidden")
}

func NotFound(w http.ResponseWriter, msg string) {
	writeError(w, "not_found", http.StatusBadRequest, msg)
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeError(w, "bad_request", http.StatusBadRequest, msg)
}

func Conflict(w http.ResponseWriter, code string, msg string) {
	writeError(w, code, http.StatusConflict, msg)
}

func Internal(w http.ResponseWriter) {
	writeError(w, "internal_error", http.StatusInternalServerError, "internal error")
}

func HandleError(w http.ResponseWriter, err error) {
	var (
		ve  *scouting.ValidationError
		nfe *scouting.NotFoundError
	)

	switch {
	case errors.Is(err, scouting.ErrAlreadyExists):
		Conflict(w, "already_exists", "resource already exists")
	case errors.As(err, &ve):
		BadRequest(w, err.Error())

		return
	case errors.As(err, &nfe):
		NotFound(w, nfe.Error())

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
