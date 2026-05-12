package lenovo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("rejects missing ClientID", func(t *testing.T) {
		if _, err := NewClient(); !errors.Is(err, ErrNoClientID) {
			t.Fatalf("err = %v, want ErrNoClientID", err)
		}
	})

	t.Run("propagates option errors", func(t *testing.T) {
		boom := errors.New("boom")
		opt := func(_ *Client) error { return boom }
		if _, err := NewClient(opt, SetClientID("x")); !errors.Is(err, boom) {
			t.Fatalf("err = %v, want %v", err, boom)
		}
	})
}

func TestSetHttpClient(t *testing.T) {
	t.Run("nil restores default", func(t *testing.T) {
		c, err := NewClient(SetClientID("x"), SetHttpClient(nil))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if c.c != http.DefaultClient {
			t.Errorf("http client = %p, want http.DefaultClient (%p)", c.c, http.DefaultClient)
		}
	})

	t.Run("custom client is used", func(t *testing.T) {
		custom := &http.Client{}
		c, err := NewClient(SetClientID("x"), SetHttpClient(custom))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if c.c != custom {
			t.Errorf("http client = %p, want %p", c.c, custom)
		}
	})
}

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

// TestRequestBuildErrors exercises the http.NewRequest error path of every
// endpoint by pointing the client at a malformed base URL containing a
// control character (which url.Parse rejects).
func TestRequestBuildErrors(t *testing.T) {
	c, err := NewClient(
		SetClientID("x"),
		SetBaseURL("http://example.com\x7f"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	cases := []struct {
		name string
		call func() error
	}{
		{"WarrantyBySerial", func() error { _, err := c.WarrantyBySerial("S"); return err }},
		{"WarrantiesBySerials", func() error { _, err := c.WarrantiesBySerials([]string{"A", "B"}); return err }},
		{"WarrantyDetailsByID", func() error { _, err := c.WarrantyDetailsByID("D"); return err }},
		{"WarrantyOptionsBySerial", func() error { _, err := c.WarrantyOptionsBySerial("DE", "S"); return err }},
		{"WarrantyOptionsByProduct", func() error { _, err := c.WarrantyOptionsByProduct("", "P"); return err }},
		{"ContractByID", func() error { _, err := c.ContractByID("C"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err == nil {
				t.Fatal("err = nil, want a non-nil request-build error")
			}
		})
	}
}

// TestTransportErrors exercises the sendRequest error path of every endpoint
// by pointing the client at a server that has already been closed, so the
// dial fails after a valid request has been constructed.
func TestTransportErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close()

	c, err := NewClient(
		SetClientID("x"),
		SetHttpClient(srv.Client()),
		SetBaseURL(url),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	cases := []struct {
		name string
		call func() error
	}{
		{"WarrantyBySerial", func() error { _, err := c.WarrantyBySerial("S"); return err }},
		{"WarrantiesBySerials", func() error { _, err := c.WarrantiesBySerials([]string{"A", "B"}); return err }},
		{"WarrantyDetailsByID", func() error { _, err := c.WarrantyDetailsByID("D"); return err }},
		{"WarrantyOptionsBySerial", func() error { _, err := c.WarrantyOptionsBySerial("DE", "S"); return err }},
		{"ContractByID", func() error { _, err := c.ContractByID("C"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err == nil {
				t.Fatal("err = nil, want transport error")
			}
		})
	}
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
