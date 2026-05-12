package lenovo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseURL     = "https://supportapi.lenovo.com/v2.5"
	warrantyURL = baseURL + "/warranty"

	InvalidCountry = "**INVALID**"

	contentTypeForm = "application/x-www-form-urlencoded"
)

// WarrantyType values returned by the Warranty Details endpoint.
const (
	WarrantyTypeBase     = "BASE"
	WarrantyTypeUpgrade  = "UPGRADE"
	WarrantyTypeExtended = "EXTENDED"
	WarrantyTypeInstant  = "INSTANT"
	WarrantyTypeUnknown  = "UNKNOWN"
)

// WarrantyDelivery values returned by the Warranty Details endpoint.
const (
	DeliveryBringIn     = "BRING_IN"
	DeliveryCourier     = "COURIER"
	DeliveryCRU         = "CRU"
	DeliveryDepot       = "DEPOT"
	DeliveryAdvExchange = "ADV_EXCHANGE"
	DeliveryOnSite      = "ON_SITE"
	DeliveryPartsOnly   = "PARTS_ONLY"
	DeliveryTechSupport = "TECH_SUPPORT"
	DeliveryUnknown     = "UNKNOWN"
)

// WarrantyCategory values returned by the Warranty Details endpoint.
const (
	CategoryMachine   = "MACHINE"
	CategoryComponent = "COMPONENT"
	CategoryUnknown   = "UNKNOWN"
)

var (
	ErrNotEnoughSerials = errors.New("not enough serials provided require at least two")
	ErrRequestFailed    = errors.New("request failed")
	ErrInvalidResponse  = errors.New("invalid response")
)

type Warranty struct {
	Serial       string
	ErrorCode    int
	ErrorMessage string
	Product      string
	InWarranty   bool
	Purchased    *Time
	Shipped      *Time
	Country      string
	UpgradeURL   string `json:"UpgradeUrl"`
	Warranty     []WarrantyWarranty
	Contract     []WarrantyContract
}

type WarrantyWarranty struct {
	ID          string
	Name        string
	Description string
	Type        string
	Start       Time
	End         Time
}

type WarrantyContract struct {
	Contract        string
	Quantity        int
	ItemNumber      string
	ChargeCode      string
	SLA             string
	EntitlementCode string
	Status          string
	Start           Time
	End             Time
}

// WarrantyDetails describes a warranty offering identified by its SDF code.
type WarrantyDetails struct {
	ID          string
	Name        string
	Description string
	Type        string
	Delivery    string
	Category    string
	Duration    string
}

func (c *Client) WarrantyBySerial(serial string) (*Warranty, error) {
	data := url.Values{}
	data.Set("Serial", serial)

	r, err := newFormRequest(http.MethodPost, warrantyURL, data)
	if err != nil {
		return nil, err
	}

	var w Warranty
	if err := c.do(r, &w); err != nil {
		return nil, err
	}
	if w.Serial == "" {
		return nil, ErrInvalidResponse
	}
	return &w, nil
}

func (c *Client) WarrantiesBySerials(serials []string) ([]Warranty, error) {
	if len(serials) <= 1 {
		return nil, ErrNotEnoughSerials
	}

	data := url.Values{}
	data.Set("Serial", serials[0])
	for _, v := range serials[1:] {
		data.Add("Serial", v)
	}

	r, err := newFormRequest(http.MethodPost, warrantyURL, data)
	if err != nil {
		return nil, err
	}

	var w []Warranty
	if err := c.do(r, &w); err != nil {
		return nil, err
	}
	return w, nil
}

// WarrantyDetailsByID looks up a warranty offering by its SDF code.
//
// See https://supportapi.lenovo.com/documentation/Warranty.html
func (c *Client) WarrantyDetailsByID(id string) (*WarrantyDetails, error) {
	r, err := http.NewRequest(http.MethodGet, warrantyURL+"/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	var w WarrantyDetails
	if err := c.do(r, &w); err != nil {
		return nil, err
	}
	if w.ID == "" {
		return nil, ErrInvalidResponse
	}
	return &w, nil
}

// WarrantyOption represents a single international warranty option for a
// destination country.
type WarrantyOption struct {
	ID          string
	Name        string
	Description string
	Type        string
	Delivery    string
	Category    string
	Duration    string
	Country     string
}

// WarrantyOptionsBySerial returns the international warranty options for the
// given serial number. countryCode may be empty to use the API default.
//
// See https://supportapi.lenovo.com/documentation/Warranty.html
func (c *Client) WarrantyOptionsBySerial(countryCode, serial string) ([]WarrantyOption, error) {
	return c.warrantyOptions(countryCode, "Serial", serial)
}

// WarrantyOptionsByProduct returns the international warranty options for the
// given product (catalog identifier). countryCode may be empty to use the API
// default.
//
// See https://supportapi.lenovo.com/documentation/Warranty.html
func (c *Client) WarrantyOptionsByProduct(countryCode, product string) ([]WarrantyOption, error) {
	return c.warrantyOptions(countryCode, "Product", product)
}

func (c *Client) warrantyOptions(countryCode, key, value string) ([]WarrantyOption, error) {
	u := baseURL + "/warrantyoption"
	if countryCode != "" {
		u += "/" + url.PathEscape(countryCode)
	}

	data := url.Values{}
	data.Set(key, value)

	r, err := newFormRequest(http.MethodPost, u, data)
	if err != nil {
		return nil, err
	}

	var opts []WarrantyOption
	if err := c.do(r, &opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// newFormRequest builds an HTTP request with a form-encoded body. The Go HTTP
// client computes the Content-Length automatically from the *strings.Reader.
func newFormRequest(method, endpoint string, data url.Values) (*http.Request, error) {
	r, err := http.NewRequest(method, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentTypeForm)
	return r, nil
}

// do sends the request, asserts a 2xx response, and decodes the body into v.
func (c *Client) do(r *http.Request, v any) error {
	resp, err := c.sendRequest(r)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
