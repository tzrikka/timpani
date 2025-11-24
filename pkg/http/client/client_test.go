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
		accept      string
		contentType string
		body        any
		wantErr     bool
	}{
		{
			name:        "get",
			startServer: true,
			httpMethod:  http.MethodGet,
			accept:      AcceptJSON,
			contentType: ContentJSON,
			body:        url.Values{},
		},
		{
			name:        "post",
			startServer: true,
			httpMethod:  http.MethodPost,
			accept:      AcceptJSON,
			contentType: ContentJSON,
			body:        "body",
		},
		{
			name:        "server_not_responding",
			httpMethod:  http.MethodPost,
			accept:      AcceptJSON,
			contentType: ContentJSON,
			body:        "body",
			wantErr:     true,
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

			got, _, err := HTTPRequest(t.Context(), tt.httpMethod, s.URL, "token", tt.accept, tt.contentType, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPRequest() error = %v, want %v", err, tt.wantErr)
				return
			}
			if tt.startServer && string(got) != want {
				t.Errorf("HTTPRequest() = %q, want %q", string(got), want)
			}
		})
	}
}

func handler(t *testing.T) http.HandlerFunc {
	t.Helper()

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

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		retryAfter int
		body       []byte
		wantErr    string
	}{
		{
			name:       "200_ok",
			statusCode: http.StatusOK,
			body:       []byte(`{"key":"value"}`),
		},
		{
			name:       "400_bad_request",
			statusCode: http.StatusBadRequest,
			wantErr:    "400 Bad Request",
		},
		{
			name:       "429_too_many_requests",
			statusCode: http.StatusTooManyRequests,
			retryAfter: 5,
			body:       []byte("retry error text"),
			wantErr:    "429 Too Many Requests (retry after 5 seconds): retry error text",
		},
		{
			name:       "429_too_many_requests_invalid_header",
			statusCode: http.StatusTooManyRequests,
			retryAfter: 0,
			body:       []byte("retry error text"),
			wantErr:    "429 Too Many Requests: retry error text",
		},
		{
			name:       "500_internal_server_error",
			statusCode: http.StatusInternalServerError,
			body:       []byte("internal server error"),
			wantErr:    "500 Internal Server Error: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Status:     fmt.Sprintf("%d %s", tt.statusCode, http.StatusText(tt.statusCode)),
				StatusCode: tt.statusCode,
			}
			if tt.retryAfter > 0 {
				resp.Header = make(http.Header)
				resp.Header.Set("Retry-After", fmt.Sprintf("%d", tt.retryAfter))
			}

			_, gotWait, err := parseResponse(resp, tt.body)
			if (err != nil) != (tt.wantErr != "") || (err != nil && err.Error() != tt.wantErr) {
				t.Errorf("parseResponse() error = %v, want %q", err, tt.wantErr)
				return
			}

			if gotWait != tt.retryAfter {
				t.Errorf("parseResponse() wait = %d, want %d", gotWait, tt.retryAfter)
			}
		})
	}
}
