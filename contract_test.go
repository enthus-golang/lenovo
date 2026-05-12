package lenovo

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestContractByID(t *testing.T) {
	const body = `{
		"ID": "CT-001", "Description": "Premier",
		"Start": "2024-01-01T00:00:00Z", "End": "2027-01-01T00:00:00Z",
		"Products": [{
			"Serial": "MP1ABCDE", "MachineType": "20HE", "Name": "ThinkPad",
			"Quantity": 1, "Status": "ACTIVE",
			"Start": "2024-01-01T00:00:00Z", "End": "2027-01-01T00:00:00Z"
		}]
	}`

	var seen http.Request
	srv := newTestServer(t, respondJSON(&seen, body))
	c := newTestClient(t, srv)

	got, err := c.ContractByID("CT-001")
	if err != nil {
		t.Fatalf("ContractByID: %v", err)
	}
	if seen.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", seen.Method)
	}
	if seen.URL.Path != "/Contract/CT-001" {
		t.Errorf("path = %q, want /Contract/CT-001", seen.URL.Path)
	}
	want := &Contract{
		ID: "CT-001", Description: "Premier",
		Start: Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		End:   Time{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		Products: []ContractProduct{{
			Serial: "MP1ABCDE", MachineType: "20HE", Name: "ThinkPad",
			Quantity: 1, Status: "ACTIVE",
			Start: Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			End:   Time{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Contract mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestContractByID_Errors(t *testing.T) {
	t.Run("non-200 returns ErrRequestFailed", func(t *testing.T) {
		srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusNotFound)
		})
		c := newTestClient(t, srv)
		if _, err := c.ContractByID("X"); !errors.Is(err, ErrRequestFailed) {
			t.Errorf("err = %v, want ErrRequestFailed", err)
		}
	})

	t.Run("empty body returns ErrInvalidResponse", func(t *testing.T) {
		srv := newTestServer(t, respondJSON(nil, `{}`))
		c := newTestClient(t, srv)
		if _, err := c.ContractByID("X"); !errors.Is(err, ErrInvalidResponse) {
			t.Errorf("err = %v, want ErrInvalidResponse", err)
		}
	})
}

func TestContractByID_EscapesPath(t *testing.T) {
	var rawPath string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"ID":"CT/01"}`)
	})
	c := newTestClient(t, srv)

	if _, err := c.ContractByID("CT/01"); err != nil {
		t.Fatalf("ContractByID: %v", err)
	}
	if want := "/Contract/CT%2F01"; rawPath != want {
		t.Errorf("escaped path = %q, want %q", rawPath, want)
	}
}
