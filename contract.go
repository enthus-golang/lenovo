package lenovo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const contractURL = baseURL + "/Contract"

// Contract represents a Lenovo service contract and the products it covers.
type Contract struct {
	ID          string
	Description string
	Start       Time
	End         Time
	Products    []ContractProduct
}

// ContractProduct is a single product covered by a Contract.
type ContractProduct struct {
	Serial          string
	MachineType     string
	Name            string
	Quantity        int
	ItemNumber      string
	ChargeCode      string
	SLA             string
	EntitlementCode string
	Status          string
	Start           Time
	End             Time
}

// ContractByID looks up the details of a service contract by its ID.
//
// See https://supportapi.lenovo.com/documentation/Warranty.html
func (c *Client) ContractByID(id string) (*Contract, error) {
	r, err := http.NewRequest(http.MethodGet, contractURL+"/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}
	if resp.ContentLength == 2 {
		return nil, ErrInvalidResponse
	}

	var contract Contract
	if err := json.NewDecoder(resp.Body).Decode(&contract); err != nil {
		return nil, err
	}
	return &contract, nil
}
