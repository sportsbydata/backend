package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-pkgz/routegroup"
	"github.com/gofrs/uuid/v5"
	"github.com/iris-contrib/schema"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/access"
	"github.com/sportsbydata/backend/scouting"
)

//go:embed static/*
var static embed.FS

type Server struct {
	sdb     *sqlx.DB
	store   scouting.Store
	decoder *schema.Decoder
	hserver *http.Server
	dev     bool

	wg      sync.WaitGroup
	closeCh chan struct{}
}

func New(sdb *sqlx.DB, store scouting.Store, addr string, dev bool) *Server {
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
		store:   store,
		decoder: dec,
		closeCh: make(chan struct{}),
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

	group.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Header.Get("Authorization"))

			h.ServeHTTP(w, r)
		})
	})

	group.Route(func(b *routegroup.Bundle) {
		b.With(withOrg).HandleFunc("POST /organization/{id}", rt.createOrganization)
		b.With(withOrg).HandleFunc("GET /account", rt.getAccounts)
		b.With(withOrg).HandleFunc("GET /league", rt.getLeagues)
		b.With(withOrgPerm(access.PermissionManageLeagues)).HandleFunc("POST /league", rt.createLeague)
		b.With(withOrgPerm(access.PermissionManageTeams)).HandleFunc("POST /team", rt.createTeam)
		b.With(withOrg).HandleFunc("GET /team", rt.getTeams)
		b.With(withOrgPerm(access.PermissionManageLeagues)).HandleFunc("PUT /organization-league", rt.updateOrganizationLeagues)
		b.HandleFunc("POST /match", rt.createMatch)
		b.HandleFunc("PATCH /match", rt.editMatch)
		b.With(withOrg).HandleFunc("GET /match", rt.getMatches)
		b.With(withOrg).HandleFunc("GET /match-scout", rt.getMatchScouts)
		b.HandleFunc("POST /match-scout", rt.createMatchScout)
		b.HandleFunc("PATCH /match-scout", rt.updateMatchScout)
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

func NotFound(w http.ResponseWriter) {
	writeError(w, "not_found", http.StatusBadRequest, "not_found")
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeError(w, "bad_request", http.StatusBadRequest, msg)
}

func Conflict(w http.ResponseWriter, msg string) {
	writeError(w, "conflict", http.StatusConflict, msg)
}

func Internal(w http.ResponseWriter) {
	writeError(w, "internal_error", http.StatusInternalServerError, "internal error")
}

func HandleError(w http.ResponseWriter, err error) {
	var ve *scouting.ValidationError

	switch {
	case errors.Is(err, scouting.ErrAlreadyExists):
		Conflict(w, "resource already exists")
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
