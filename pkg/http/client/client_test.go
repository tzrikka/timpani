package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPRequest(t *testing.T) {
	tests := []struct {
		name        string
		startServer bool
		httpMethod  string
		body        any
		wantErr     bool
	}{
		{
			name:        "get",
			startServer: true,
			httpMethod:  http.MethodGet,
			body:        url.Values{},
		},
		{
			name:        "post",
			startServer: true,
			httpMethod:  http.MethodPost,
			body:        "body",
		},
		{
			name:       "server_not_responding",
			httpMethod: http.MethodPost,
			body:       "body",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		want := "body\n"
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			got, err := HTTPRequest(t.Context(), tt.httpMethod, s.URL, "token", tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.startServer && string(got) != want {
				t.Errorf("HTTPRequest() = %q, want %q", string(got), want)
			}
		})
	}
}

func handler(t *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Accept")
		want := "application/json"
		if got != want {
			t.Errorf("accept header = %q, want %q", got, want)
		}

		got = r.Header.Get("Authorization")
		want = "Bearer token"
		if got != want {
			t.Errorf("authorization header = %q, want %q", got, want)
		}

		body := "body\n"
		n, err := fmt.Fprint(w, body)
		if err != nil {
			t.Errorf("failed to write body: %v", err)
		}
		if n != len(body) {
			t.Errorf("wrote %d body bytes, want %d", n, len(body))
		}
	})
}
