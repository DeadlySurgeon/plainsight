package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var w = fmt.Errorf

type ClientOpt func(*clientOpts)

type clientOpts struct {
	baseURL  string
	username string
	password string

	client *http.Client
}

func ServiceURL(url string) func(*clientOpts) {
	return func(co *clientOpts) {
		co.baseURL = url
	}
}

func BasicAuth(username, password string) func(*clientOpts) {
	return func(co *clientOpts) {
		co.username = username
		co.password = password
	}
}

func WithHTTPClient(client *http.Client) func(*clientOpts) {
	return func(co *clientOpts) {
		co.client = client
	}
}

type client struct {
	clientOpts
}

var ErrInvalidClientOpts = errors.New("invalid client opts")
var ErrNoUsernameProvided = fmt.Errorf("%w: no username provided", ErrInvalidClientOpts)
var ErrNoPasswordProvided = fmt.Errorf("%w: no password provided", ErrInvalidClientOpts)
var ErrInvalidURLProvided = errors.New("invalid url")

func NewClient(opts ...ClientOpt) (*client, error) {
	cfg := clientOpts{baseURL: "", username: "", password: ""}
	// Go over each option, apply them to the base configuration
	for _, opt := range opts {
		opt(&cfg)
	}

	// Username is required
	if cfg.username == "" {
		return nil, ErrNoUsernameProvided
	}

	// Password is required
	if cfg.password == "" {
		return nil, ErrNoPasswordProvided
	}

	// Check to see if the base URL is provided.
	if cfg.baseURL == "" {
		return nil, w("%w: not provided", ErrInvalidURLProvided)
	}

	// Ensure the base URL is good to go.
	u, err := url.Parse(clean(cfg.baseURL))
	if err != nil {
		return nil, w("%w: %w", ErrInvalidURLProvided, err)
	}
	// After cleaning the base url, put it back in place. This ensures that
	// anything bad that might have been in it that we don't want is stripped
	// out leaving just the url.
	cfg.baseURL = u.String()

	// Check to see if the client was provided.
	if cfg.client == nil {
		// Default sane client
		cfg.client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	// Return the client with the configurations
	return &client{cfg}, nil
}

// Clean is needed due to part of the hypervisor configurator sometimes giving
// bad input that is otherwise still ASCII. This ensures that the string is
// still usable. Legacy systems, am I right?
func clean(s string) string {
	var b []byte
	for _, c := range s {
		b = append(b, byte(c))
	}
	return string(b)
}

var ErrUnableToFormRequest = errors.New("unable to form request")
var ErrFailedToPerformRequest = errors.New("failed to perform request")
var ErrMalformedServerData = errors.New("malformed server data")

type ErrRequestFailure struct {
	baseStatus int
}

func (err *ErrRequestFailure) Status() int { return err.baseStatus }
func (err *ErrRequestFailure) Error() string {
	return "request returned an unexpected status code of " + strconv.Itoa(err.baseStatus)
}

// RequestJWT uses the provided basic auth to request a JWT. Not ideal, but this
// could be a legacy auth system that can only do basic auth.
func (c *client) RequestJWT(ctx context.Context) (string, error) {
	fmt.Println("Making request to", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return "", w("%w: %w", ErrUnableToFormRequest, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", w("%w: %w", ErrFailedToPerformRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &ErrRequestFailure{resp.StatusCode}
	}

	var jwtResp jwtResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&jwtResp); err != nil {
		return "", w("%w: %w", ErrMalformedServerData, err)
	}

	return jwtResp.Token, nil
}

type jwtResponsePayload struct {
	Token string `json:"token"`
}
