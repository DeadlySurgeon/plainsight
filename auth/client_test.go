package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		opts     []ClientOpt
		wantErr  error
		wantURL  string
		wantHTTP bool
	}{
		{
			name:    "missing username",
			opts:    []ClientOpt{BasicAuth("", "password"), ServiceURL("http://provider.cluster.local")},
			wantErr: ErrNoUsernameProvided,
		},
		{
			name:    "missing password",
			opts:    []ClientOpt{BasicAuth("user", ""), ServiceURL("http://provider.cluster.local")},
			wantErr: ErrNoPasswordProvided,
		},
		{
			name:    "empty URL",
			opts:    []ClientOpt{BasicAuth("user", "pass"), ServiceURL("")},
			wantErr: ErrInvalidURLProvided,
		},
		{
			name:    "invalid URL",
			opts:    []ClientOpt{BasicAuth("user", "pass"), ServiceURL("http://\x7f/")},
			wantErr: ErrInvalidURLProvided,
		},
		{
			name:     "valid opts with default http client",
			opts:     []ClientOpt{BasicAuth("user", "pass"), ServiceURL("http://provider.cluster.local")},
			wantErr:  nil,
			wantURL:  "http://provider.cluster.local",
			wantHTTP: true,
		},
		{
			name: "valid opts with custom http client",
			opts: []ClientOpt{
				BasicAuth("user", "pass"),
				ServiceURL("http://provider.cluster.local"),
				WithHTTPClient(&http.Client{}),
			},
			wantErr:  nil,
			wantURL:  "http://provider.cluster.local",
			wantHTTP: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.opts...)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.baseURL != tt.wantURL {
				t.Errorf("expected baseURL %q, got %q", tt.wantURL, c.baseURL)
			}
			if tt.wantHTTP && c.client == nil {
				t.Errorf("expected http client to be set, got nil")
			}
		})
	}
}

func TestRequestJWT(t *testing.T) {
	tests := map[string]struct {
		serverStatus int
		serverBody   string
		serverClosed bool

		ctx       context.Context
		wantToken string
		wantErr   error
	}{
		"success": {
			serverStatus: http.StatusOK,
			serverBody:   `{"token":"abc123"}`,

			ctx:       context.Background(),
			wantToken: "abc123",
		},
		"nil context": {
			ctx:     nil,
			wantErr: ErrUnableToFormRequest,
		},
		"network failure": {
			serverClosed: true,
			ctx:          context.Background(),
			wantErr:      ErrFailedToPerformRequest,
		},
		"bad status code": {
			serverStatus: http.StatusForbidden,
			ctx:          context.Background(),
			wantErr: &ErrRequestFailure{
				baseStatus: http.StatusForbidden,
			},
		},
		"malformed JSON": {
			serverBody: `{"bad_token`,
			ctx:        context.Background(),
			wantErr:    ErrMalformedServerData,
		},
		"bad json": {
			serverBody: `{"token": 1234}`,
			ctx:        context.Background(),
			wantErr:    ErrMalformedServerData,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.serverStatus == 0 {
					test.serverStatus = http.StatusOK
				}
				w.WriteHeader(test.serverStatus)

				_, _ = w.Write([]byte(test.serverBody))
			}))
			if srv != nil {
				if test.serverClosed {
					srv.Close()
				} else {
					defer srv.Close()
				}
			}

			url := ""
			if srv != nil {
				url = srv.URL
			}

			c, err := NewClient(
				BasicAuth("user", "pass"),
				ServiceURL(url),
				WithHTTPClient(&http.Client{}),
			)
			if err != nil {
				t.Fatalf("unexpected error creating client: %v", err)
			}

			token, err := c.RequestJWT(test.ctx)

			if test.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %T, got nil", test.wantErr)
				}
				switch test.wantErr.(type) {
				case *ErrRequestFailure:
					var reqFail *ErrRequestFailure
					if !errors.As(err, &reqFail) {
						t.Fatalf("expected ErrRequestFailure, got %v", err)
					}
					if reqFail.Error() == "" || reqFail.Status() != test.serverStatus {
						t.Fatal("failed to capture status report: %w", err)
					}
				default:
					if !errors.Is(err, test.wantErr) {
						t.Fatalf("expected error %v, got %v", test.wantErr, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != test.wantToken {
				t.Errorf("expected token %q, got %q", test.wantToken, token)
			}
		})
	}
}
