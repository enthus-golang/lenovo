package lenovo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	warrantyURL = "https://supportapi.lenovo.com/v2.5/warranty"

	InvalidCountry = "**INVALID**"
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

func (c *Client) WarrantyBySerial(serial string) (*Warranty, error) {
	data := url.Values{}
	data.Set("Serial", serial)

	dataEncoded := data.Encode()
	r, err := http.NewRequest(http.MethodPost, warrantyURL, strings.NewReader(dataEncoded))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(dataEncoded)))

	resp, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}
	if resp.ContentLength == 2 {
		return nil, ErrInvalidResponse
	}

	var w Warranty
	err = json.NewDecoder(resp.Body).Decode(&w)
	if err != nil {
		return nil, err
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

	dataEncoded := data.Encode()
	r, err := http.NewRequest(http.MethodPost, warrantyURL, strings.NewReader(dataEncoded))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(dataEncoded)))

	resp, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}

	var w []Warranty
	err = json.NewDecoder(resp.Body).Decode(&w)
	if err != nil {
		return nil, err
	}

	return w, nil
}
