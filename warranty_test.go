package lenovo

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestWarrantyBySerial(t *testing.T) {
	const body = `{
		"Serial": "MP1ABCDE",
		"Product": "20HES03P00",
		"InWarranty": true,
		"Purchased": "2020-04-15T00:00:00Z",
		"Shipped": "2020-04-20T00:00:00Z",
		"Country": "DE",
		"UpgradeUrl": "https://example.com/upgrade",
		"Warranty": [
			{
				"ID": "3Y-DEPOT", "Name": "3 Year Depot", "Description": "Depot warranty",
				"Type": "BASE",
				"Start": "2020-04-20T00:00:00Z", "End": "2023-04-20T00:00:00Z"
			}
		],
		"Contract": []
	}`

	var seen http.Request
	srv := newTestServer(t, respondJSON(&seen, body))
	c := newTestClient(t, srv)

	got, err := c.WarrantyBySerial("MP1ABCDE")
	if err != nil {
		t.Fatalf("WarrantyBySerial: %v", err)
	}

	// Request shape.
	if seen.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", seen.Method)
	}
	if seen.URL.Path != "/warranty" {
		t.Errorf("path = %q, want /warranty", seen.URL.Path)
	}
	if h := seen.Header.Get("ClientID"); h != "test-client-id" {
		t.Errorf("ClientID = %q, want test-client-id", h)
	}
	if h := seen.Header.Get("Content-Type"); h != contentTypeForm {
		t.Errorf("Content-Type = %q, want %q", h, contentTypeForm)
	}
	if s := seen.PostForm.Get("Serial"); s != "MP1ABCDE" {
		t.Errorf("Serial form = %q, want MP1ABCDE", s)
	}
	if int64(len(seen.PostForm.Encode())) != seen.ContentLength {
		t.Errorf("ContentLength = %d, form bytes = %d", seen.ContentLength, len(seen.PostForm.Encode()))
	}

	// Response decoding.
	want := &Warranty{
		Serial:     "MP1ABCDE",
		Product:    "20HES03P00",
		InWarranty: true,
		Purchased:  &Time{time.Date(2020, 4, 15, 0, 0, 0, 0, time.UTC)},
		Shipped:    &Time{time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)},
		Country:    "DE",
		UpgradeURL: "https://example.com/upgrade",
		Warranty: []WarrantyWarranty{{
			ID: "3Y-DEPOT", Name: "3 Year Depot", Description: "Depot warranty",
			Type:  WarrantyTypeBase,
			Start: Time{time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)},
			End:   Time{time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC)},
		}},
		Contract: []WarrantyContract{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Warranty mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestWarrantyBySerial_Errors(t *testing.T) {
	t.Run("non-200 returns ErrRequestFailed", func(t *testing.T) {
		srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusInternalServerError)
		})
		c := newTestClient(t, srv)
		if _, err := c.WarrantyBySerial("X"); !errors.Is(err, ErrRequestFailed) {
			t.Errorf("err = %v, want ErrRequestFailed", err)
		}
	})

	t.Run("empty body returns ErrInvalidResponse", func(t *testing.T) {
		srv := newTestServer(t, respondJSON(nil, `{}`))
		c := newTestClient(t, srv)
		if _, err := c.WarrantyBySerial("X"); !errors.Is(err, ErrInvalidResponse) {
			t.Errorf("err = %v, want ErrInvalidResponse", err)
		}
	})
}

func TestWarrantiesBySerials(t *testing.T) {
	t.Run("encodes all serials", func(t *testing.T) {
		var seen http.Request
		srv := newTestServer(t, respondJSON(&seen, `[{"Serial":"A"},{"Serial":"B"},{"Serial":"C"}]`))
		c := newTestClient(t, srv)

		got, err := c.WarrantiesBySerials([]string{"A", "B", "C"})
		if err != nil {
			t.Fatalf("WarrantiesBySerials: %v", err)
		}
		if want := []string{"A", "B", "C"}; !reflect.DeepEqual(seen.PostForm["Serial"], want) {
			t.Errorf("Serial form values = %q, want %q", seen.PostForm["Serial"], want)
		}
		if len(got) != 3 {
			t.Errorf("len = %d, want 3", len(got))
		}
	})

	t.Run("rejects single serial", func(t *testing.T) {
		c, err := NewClient(SetClientID("x"))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		if _, err := c.WarrantiesBySerials([]string{"only-one"}); !errors.Is(err, ErrNotEnoughSerials) {
			t.Errorf("err = %v, want ErrNotEnoughSerials", err)
		}
	})
}

func TestWarrantyDetailsByID(t *testing.T) {
	const body = `{
		"ID": "3Y-DEPOT", "Name": "3 Year Depot", "Description": "Depot warranty",
		"Type": "BASE", "Delivery": "DEPOT", "Category": "MACHINE", "Duration": "3Y"
	}`

	var seen http.Request
	srv := newTestServer(t, respondJSON(&seen, body))
	c := newTestClient(t, srv)

	got, err := c.WarrantyDetailsByID("3Y-DEPOT")
	if err != nil {
		t.Fatalf("WarrantyDetailsByID: %v", err)
	}
	if seen.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", seen.Method)
	}
	if seen.URL.Path != "/warranty/3Y-DEPOT" {
		t.Errorf("path = %q, want /warranty/3Y-DEPOT", seen.URL.Path)
	}
	want := &WarrantyDetails{
		ID: "3Y-DEPOT", Name: "3 Year Depot", Description: "Depot warranty",
		Type: WarrantyTypeBase, Delivery: DeliveryDepot, Category: CategoryMachine, Duration: "3Y",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WarrantyDetails mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestWarrantyDetailsByID_Errors(t *testing.T) {
	t.Run("empty body returns ErrInvalidResponse", func(t *testing.T) {
		srv := newTestServer(t, respondJSON(nil, `{}`))
		c := newTestClient(t, srv)
		if _, err := c.WarrantyDetailsByID("X"); !errors.Is(err, ErrInvalidResponse) {
			t.Errorf("err = %v, want ErrInvalidResponse", err)
		}
	})
}

func TestWarrantyDetailsByID_EscapesPath(t *testing.T) {
	var rawPath string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"ID":"weird/value","Name":"x"}`)
	})
	c := newTestClient(t, srv)

	if _, err := c.WarrantyDetailsByID("weird/value"); err != nil {
		t.Fatalf("WarrantyDetailsByID: %v", err)
	}
	if want := "/warranty/weird%2Fvalue"; rawPath != want {
		t.Errorf("escaped path = %q, want %q", rawPath, want)
	}
}

func TestWarrantyOptions(t *testing.T) {
	cases := []struct {
		name     string
		call     func(*Client) ([]WarrantyOption, error)
		wantPath string
		wantKey  string
		wantVal  string
	}{
		{
			name:     "by serial with country",
			call:     func(c *Client) ([]WarrantyOption, error) { return c.WarrantyOptionsBySerial("US", "MP1ABCDE") },
			wantPath: "/warrantyoption/US",
			wantKey:  "Serial", wantVal: "MP1ABCDE",
		},
		{
			name:     "by product without country",
			call:     func(c *Client) ([]WarrantyOption, error) { return c.WarrantyOptionsByProduct("", "20HES03P00") },
			wantPath: "/warrantyoption",
			wantKey:  "Product", wantVal: "20HES03P00",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen http.Request
			srv := newTestServer(t, respondJSON(&seen, `[]`))
			c := newTestClient(t, srv)

			if _, err := tc.call(c); err != nil {
				t.Fatalf("call: %v", err)
			}
			if seen.URL.Path != tc.wantPath {
				t.Errorf("path = %q, want %q", seen.URL.Path, tc.wantPath)
			}
			if got := seen.PostForm.Get(tc.wantKey); got != tc.wantVal {
				t.Errorf("%s form = %q, want %q", tc.wantKey, got, tc.wantVal)
			}
		})
	}
}
