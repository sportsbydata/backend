package server

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func withOrgPerm(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return clerkhttp.WithHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				Unauthorized(w)

				return
			}

			if claims.ActiveOrganizationID == "" {
				Unauthorized(w)

				return
			}

			if !claims.HasPermission(perm) {
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

		if claims.ActiveOrganizationID == "" {
			Unauthorized(w)

			return
		}

		next.ServeHTTP(w, r)
	}))
}

func withCorsBypass(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Authorization,Content-Type")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}
