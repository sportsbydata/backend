package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func withRouteMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

func withBasicAuth(key []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, pwd, ok := r.BasicAuth()
			if !ok {
				slog.Warn("attempted to auth without key")
				w.WriteHeader(http.StatusNotFound)

				return
			}

			enc := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, pwd)))

			if bytes.Equal(key, []byte(enc)) {
				slog.Warn("attempted to auth with invalid credetnials", slog.String("username", username))
				w.WriteHeader(http.StatusNotFound)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func withOrgPerm(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return clerkhttp.WithHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				Unauthorized(w)

				return
			}

			logger := slog.With(
				slog.String("pattern", r.Pattern),
				slog.String("permission", perm),
			)

			if claims.ActiveOrganizationID == "" {
				logger.Warn("unauthorized without active organization id")
				Unauthorized(w)

				return
			}

			if !claims.HasPermission(perm) {
				logger.Warn("unauthorized without permission")
				Forbidden(w)

				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

func withOrg(next http.Handler) http.Handler {
	return clerkhttp.WithHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok {
			Unauthorized(w)

			return
		}

		logger := slog.With(slog.String("pattern", r.Pattern))

		if claims.ActiveOrganizationID == "" {
			logger.Warn("unauthorized without organization")
			Unauthorized(w)

			return
		}

		next.ServeHTTP(w, r)
	}))
}
