package lenovo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer spins up an httptest.Server and registers t.Cleanup to shut
// it down.
func newTestServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// newTestClient builds a Client pointed at srv.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(
		SetClientID("test-client-id"),
		SetHttpClient(srv.Client()),
		SetBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// respondJSON returns a handler that parses the form, optionally records the
// request into seen, and writes body as JSON.
func respondJSON(seen *http.Request, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if seen != nil {
			*seen = *r
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}
}
