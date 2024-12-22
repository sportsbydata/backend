package router

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

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
