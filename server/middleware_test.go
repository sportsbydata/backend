package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_withBasicAuth(t *testing.T) {
	hdl := withBasicAuth("Token")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	t.Run("correct token", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "Basic Token")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("invalid capitilization", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "Basic token")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "Basic token2")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("missing Basic keyword", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "token")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
