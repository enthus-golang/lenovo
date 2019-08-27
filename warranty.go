package lenovo

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	warrantyURL = "https://supportapi.lenovo.com/v2.5/warranty"

	InvalidCountry = "**INVALID**"
)

var (
	ErrNotEnoughSerials = errors.New("not enough serials provided require at least two")
	ErrRequestFailed    = errors.New("request failed")
)

type Serial struct {
	Serial       string
	ErrorCode    int
	ErrorMessage string
	Product      string
	InWarranty   bool
	Purchased    *Time
	Shipped      *Time
	Country      string
	UpgradeURL   string `json:"UpgradeUrl"`
	Warranty     []SerialWarranty
}

type SerialWarranty struct {
	ID          string
	Name        string
	Description string
	Type        string
	Start       Time
	End         Time
}

func (c *Client) WarrantyBySerial(serial string) (*Serial, error) {
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
		return nil, errors.Wrap(ErrRequestFailed, resp.Status)
	}

	var s Serial
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (c *Client) WarrantiesBySerials(serials []string) ([]Serial, error) {
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
		return nil, errors.Wrap(ErrRequestFailed, resp.Status)
	}

	var s []Serial
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
