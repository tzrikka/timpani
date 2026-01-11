package webhooks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestHTTPServerThrippyHandler(t *testing.T) {
	tests := []struct {
		name        string
		reqMethod   string
		reqURL      string
		wantQuery   url.Values
		thrippyResp *http.Response
		wantResp    *http.Response
	}{
		{
			name:      "no_thrippy_server",
			reqMethod: http.MethodGet,
			reqURL:    "https://example.com",
			wantResp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{},
			},
		},
		{
			name:      "ok",
			reqMethod: http.MethodGet,
			reqURL:    "https://example.com",
			thrippyResp: &http.Response{
				StatusCode: http.StatusOK,
			},
			wantResp: &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Length": []string{"0"},
				},
			},
		},
		{
			name:      "accepted_with_data",
			reqMethod: http.MethodGet,
			reqURL:    "https://example.com?k1=v1&k2=v2",
			wantQuery: map[string][]string{"k1": {"v1"}, "k2": {"v2"}},
			thrippyResp: &http.Response{
				StatusCode: http.StatusAccepted,
				Header: http.Header{
					"X-Test-Header": []string{"3", "4"},
				},
			},
			wantResp: &http.Response{
				StatusCode: http.StatusAccepted,
				Header: http.Header{
					"Content-Length": []string{"0"},
					"X-Test-Header":  []string{"3", "4"},
				},
			},
		},
		{
			name:      "redirect",
			reqMethod: http.MethodGet,
			reqURL:    "https://example.com/redirect",
			thrippyResp: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
			wantResp: &http.Response{
				StatusCode: http.StatusFound,
				Header: http.Header{
					"Content-Length": []string{"24"},
					"Content-Type":   []string{"text/html; charset=utf-8"},
					"Location":       []string{"/"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock Thrippy server.
			s := httptest.NewUnstartedServer(mockThrippyServer(t, tt.wantQuery, tt.thrippyResp))
			thrippyBaseURL := "http://127.0.0.1:0"
			if tt.thrippyResp != nil {
				s.Start()
				thrippyBaseURL = s.URL
			}
			defer s.Close()

			u, _ := url.Parse(thrippyBaseURL)
			server := &HTTPServer{thrippyURL: u}

			// Construct client request.
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(t.Context(), tt.reqMethod, tt.reqURL, http.NoBody)
			r.Header.Add("X-Test-Header", "1")
			r.Header.Add("X-Test-Header", "2")

			// Send client request to handler under test.
			server.thrippyHandler(w, r)
			got := w.Result()

			// Response checks.
			if got.StatusCode != tt.wantResp.StatusCode {
				t.Errorf("response status code: got %v, want %v", got.StatusCode, tt.wantResp.StatusCode)
			}

			got.Header.Del("Date")
			if !reflect.DeepEqual(got.Header, tt.wantResp.Header) {
				t.Errorf("response headers: got %v, want %v", got.Header, tt.wantResp.Header)
			}
		})
	}
}

func mockThrippyServer(t *testing.T, wantQuery url.Values, resp *http.Response) http.HandlerFunc {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Support redirection in tests.
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Request checks.
		hs := r.Header.Values("X-Test-Header")
		if len(hs) != 2 || hs[0] != "1" || hs[1] != "2" {
			t.Errorf("thrippy request headers: got %v, want [1 2]", hs)
		}

		if wantQuery == nil {
			wantQuery = url.Values{}
		}
		if gotQuery := r.URL.Query(); !reflect.DeepEqual(gotQuery, wantQuery) {
			t.Errorf("thrippy request query: got %v, want %v", gotQuery, wantQuery)
		}

		// Construct response.
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
	})
}
