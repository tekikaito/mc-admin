package ashcon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAshconClient_CheckMojangUsernameExists(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantExists bool
		wantErr    bool
		wantErrSub string
	}{
		{name: "200 means exists", statusCode: 200, wantExists: true},
		{name: "404 means does not exist", statusCode: 404, wantExists: false},
		{name: "500 means error", statusCode: 500, wantErr: true, wantErrSub: "unexpected response code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer srv.Close()

			c := &AshconClient{apiURL: srv.URL + "/"}

			got, err := c.CheckMojangUsernameExists("Steve")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrSub != "" && !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("error = %q, expected to contain %q", err.Error(), tt.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantExists {
				t.Fatalf("exists = %v, want %v", got, tt.wantExists)
			}
		})
	}
}

func TestAshconClient_CheckMojangUsernameExists_networkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	// Close immediately so requests fail.
	srv.Close()

	c := &AshconClient{apiURL: srv.URL + "/"}

	_, err := c.CheckMojangUsernameExists("Steve")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
