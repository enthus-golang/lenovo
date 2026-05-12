package lenovo

import (
	"net/http"
	"net/url"
)

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
	r, err := http.NewRequest(http.MethodGet, c.baseURL+"/Contract/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}

	var contract Contract
	if err := c.do(r, &contract); err != nil {
		return nil, err
	}
	if contract.ID == "" {
		return nil, ErrInvalidResponse
	}
	return &contract, nil
}
