package lenovo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

var (
	ErrNoClientID = errors.New("no ClientID set")
)

type ClientOptionFunc func(*Client) error

type Client struct {
	c  *http.Client
	id string
}

// NewClient creates a new client to work with Lenovo eSupport.
func NewClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{
		c: http.DefaultClient,
	}

	// Run the options on it
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	c.c.Jar = jar

	if c.id == "" {
		return nil, ErrNoClientID
	}

	return c, nil
}

// SetHttpClient can be used to specify the http.Client to use when making
// HTTP requests to Lenovo eSupport.
func SetHttpClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *Client) error {
		if httpClient != nil {
			c.c = httpClient
		} else {
			c.c = http.DefaultClient
		}
		return nil
	}
}

// SetClientID defines the ClientID which is needed to authenticate with
// Lenovo eSupport.
func SetClientID(id string) ClientOptionFunc {
	return func(c *Client) error {
		c.id = id
		fmt.Println(c.id)
		return nil
	}
}

// sendRequest sends a http.Request and append the ClientID for
// authentication.
func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add("ClientID", c.id)

	return c.c.Do(req)
}
