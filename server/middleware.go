package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func withBasicAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Basic")
			if got == "" {
				NotFound(w, "not found")

				return
			}

			if got != token {
				NotFound(w, "not found")

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
