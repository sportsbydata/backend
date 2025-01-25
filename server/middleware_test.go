package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_withApiKey(t *testing.T) {
	hdl := withApiKey("Key")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	t.Run("correct key", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "ApiKey Key")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("invalid capitilization", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "ApiKey key")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid key", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "ApiKey key2")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("missing ApiKey keyword", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.Header.Set("Authorization", "key")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
