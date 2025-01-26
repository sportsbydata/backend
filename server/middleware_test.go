package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_withBasicAuth(t *testing.T) {
	// user:test
	key := []byte("dWVyOnRlc3QK")

	hdl := withBasicAuth(key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	t.Run("correct key", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "http://test.com/metrics", http.NoBody)
		req.SetBasicAuth("user", "test")

		hdl.ServeHTTP(rec, req)

		resp := rec.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
