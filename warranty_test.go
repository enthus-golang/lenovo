package lenovo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient wires a Client at the given httptest.Server.
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
				"ID": "3Y-DEPOT",
				"Name": "3 Year Depot",
				"Description": "Depot warranty",
				"Type": "BASE",
				"Start": "2020-04-20T00:00:00Z",
				"End": "2023-04-20T00:00:00Z"
			}
		],
		"Contract": []
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/warranty" {
			t.Errorf("path = %q, want /warranty", r.URL.Path)
		}
		if got := r.Header.Get("ClientID"); got != "test-client-id" {
			t.Errorf("ClientID header = %q, want test-client-id", got)
		}
		if got := r.Header.Get("Content-Type"); got != contentTypeForm {
			t.Errorf("Content-Type = %q, want %q", got, contentTypeForm)
		}
		// http.NewRequest with *strings.Reader auto-sets ContentLength; the
		// body byte count must match.
		bodyLen := r.ContentLength
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if int64(len(r.PostForm.Encode())) != bodyLen {
			t.Errorf("ContentLength = %d, form bytes = %d", bodyLen, len(r.PostForm.Encode()))
		}
		if got := r.PostForm.Get("Serial"); got != "MP1ABCDE" {
			t.Errorf("Serial form value = %q, want MP1ABCDE", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.WarrantyBySerial("MP1ABCDE")
	if err != nil {
		t.Fatalf("WarrantyBySerial: %v", err)
	}
	if got.Serial != "MP1ABCDE" {
		t.Errorf("Serial = %q, want MP1ABCDE", got.Serial)
	}
	if !got.InWarranty {
		t.Error("InWarranty = false, want true")
	}
	if got.Country != "DE" {
		t.Errorf("Country = %q, want DE", got.Country)
	}
	if got.UpgradeURL != "https://example.com/upgrade" {
		t.Errorf("UpgradeURL = %q", got.UpgradeURL)
	}
	if got.Purchased == nil || !got.Purchased.Equal(time.Date(2020, 4, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Purchased = %v, want 2020-04-15", got.Purchased)
	}
	if len(got.Warranty) != 1 {
		t.Fatalf("len(Warranty) = %d, want 1", len(got.Warranty))
	}
	if got.Warranty[0].Type != WarrantyTypeBase {
		t.Errorf("Warranty[0].Type = %q, want %q", got.Warranty[0].Type, WarrantyTypeBase)
	}
}

func TestWarrantyBySerialEmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.WarrantyBySerial("X")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err = %v, want ErrInvalidResponse", err)
	}
}

func TestWarrantyBySerialNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.WarrantyBySerial("X")
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("err = %v, want ErrRequestFailed", err)
	}
}

func TestWarrantiesBySerials(t *testing.T) {
	const body = `[
		{"Serial":"A","Product":"P1","InWarranty":true},
		{"Serial":"B","Product":"P2","InWarranty":false}
	]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		serials := r.PostForm["Serial"]
		want := []string{"A", "B", "C"}
		if len(serials) != len(want) {
			t.Fatalf("Serial count = %d, want %d", len(serials), len(want))
		}
		for i, s := range want {
			if serials[i] != s {
				t.Errorf("Serial[%d] = %q, want %q", i, serials[i], s)
			}
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.WarrantiesBySerials([]string{"A", "B", "C"})
	if err != nil {
		t.Fatalf("WarrantiesBySerials: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Serial != "A" || got[1].Serial != "B" {
		t.Errorf("serials = %q,%q", got[0].Serial, got[1].Serial)
	}
}

func TestWarrantiesBySerialsTooFew(t *testing.T) {
	c, err := NewClient(SetClientID("x"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.WarrantiesBySerials([]string{"only-one"}); !errors.Is(err, ErrNotEnoughSerials) {
		t.Fatalf("err = %v, want ErrNotEnoughSerials", err)
	}
}

func TestWarrantyDetailsByID(t *testing.T) {
	const body = `{
		"ID": "3Y-DEPOT",
		"Name": "3 Year Depot",
		"Description": "Depot warranty",
		"Type": "BASE",
		"Delivery": "DEPOT",
		"Category": "MACHINE",
		"Duration": "3Y"
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/warranty/3Y-DEPOT" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.WarrantyDetailsByID("3Y-DEPOT")
	if err != nil {
		t.Fatalf("WarrantyDetailsByID: %v", err)
	}
	if got.ID != "3Y-DEPOT" {
		t.Errorf("ID = %q", got.ID)
	}
	if got.Type != WarrantyTypeBase || got.Delivery != DeliveryDepot || got.Category != CategoryMachine {
		t.Errorf("enums = %q/%q/%q", got.Type, got.Delivery, got.Category)
	}
}

func TestWarrantyDetailsByIDEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.WarrantyDetailsByID("missing"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err = %v, want ErrInvalidResponse", err)
	}
}

func TestWarrantyDetailsByIDEscapesPath(t *testing.T) {
	var rawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"ID":"weird/value","Name":"x"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.WarrantyDetailsByID("weird/value"); err != nil {
		t.Fatalf("WarrantyDetailsByID: %v", err)
	}
	if rawPath != "/warranty/weird%2Fvalue" {
		t.Errorf("escaped path = %q, want /warranty/weird%%2Fvalue", rawPath)
	}
}

func TestWarrantyOptionsBySerial(t *testing.T) {
	const body = `[
		{"ID":"5Y-ONSITE","Name":"5Y","Type":"UPGRADE","Delivery":"ON_SITE","Category":"MACHINE","Country":"US"}
	]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/warrantyoption/US" {
			t.Errorf("path = %q, want /warrantyoption/US", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.PostForm.Get("Serial"); got != "MP1ABCDE" {
			t.Errorf("Serial = %q, want MP1ABCDE", got)
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.WarrantyOptionsBySerial("US", "MP1ABCDE")
	if err != nil {
		t.Fatalf("WarrantyOptionsBySerial: %v", err)
	}
	if len(got) != 1 || got[0].Country != "US" || got[0].Delivery != DeliveryOnSite {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestWarrantyOptionsByProductDefaultCountry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/warrantyoption" {
			t.Errorf("path = %q, want /warrantyoption (no country segment)", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.PostForm.Get("Product"); got != "20HES03P00" {
			t.Errorf("Product = %q", got)
		}
		_, _ = io.WriteString(w, `[]`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.WarrantyOptionsByProduct("", "20HES03P00")
	if err != nil {
		t.Fatalf("WarrantyOptionsByProduct: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestContractByID(t *testing.T) {
	const body = `{
		"ID": "CT-001",
		"Description": "Premier",
		"Start": "2024-01-01T00:00:00Z",
		"End": "2027-01-01T00:00:00Z",
		"Products": [
			{
				"Serial": "MP1ABCDE",
				"MachineType": "20HE",
				"Name": "ThinkPad",
				"Quantity": 1,
				"Status": "ACTIVE",
				"Start": "2024-01-01T00:00:00Z",
				"End": "2027-01-01T00:00:00Z"
			}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/Contract/CT-001" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ContractByID("CT-001")
	if err != nil {
		t.Fatalf("ContractByID: %v", err)
	}
	if got.ID != "CT-001" || len(got.Products) != 1 || got.Products[0].MachineType != "20HE" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestContractByIDNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.ContractByID("X"); !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("err = %v, want ErrRequestFailed", err)
	}
}

func TestContractByIDEscapesPath(t *testing.T) {
	var rawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"ID":"CT/01"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.ContractByID("CT/01"); err != nil {
		t.Fatalf("ContractByID: %v", err)
	}
	if rawPath != "/Contract/CT%2F01" {
		t.Errorf("escaped path = %q, want /Contract/CT%%2F01", rawPath)
	}
}

func TestContractByIDEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.ContractByID("missing"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err = %v, want ErrInvalidResponse", err)
	}
}
