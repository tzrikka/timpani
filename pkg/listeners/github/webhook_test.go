package github

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/tzrikka/timpani/internal/listeners"
)

func TestCheckContentTypeHeader(t *testing.T) {
	tests := []struct {
		name string
		ct   string
		want int
	}{
		{
			name: "none",
			want: http.StatusBadRequest,
		},
		{
			name: "html",
			ct:   "text/html",
			want: http.StatusBadRequest,
		},
		{
			name: "json",
			ct:   "application/json",
			want: http.StatusOK,
		},
		{
			name: "text",
			ct:   "text/plain",
			want: http.StatusBadRequest,
		},
		{
			name: "web_form",
			ct:   "application/x-www-form-urlencoded",
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := listeners.RequestData{
				Headers: http.Header{
					contentTypeHeader: []string{tt.ct},
				},
			}

			if got := checkContentTypeHeader(slog.Default(), r); got != tt.want {
				t.Errorf("checkContentTypeHeader() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckSignatureHeader(t *testing.T) {
	tests := []struct {
		name   string
		sig    string
		secret string
		want   int
	}{
		{
			name: "none",
			want: http.StatusForbidden,
		},
		{
			name: "signing_secret_not_configured",
			sig:  "hash",
			want: http.StatusInternalServerError,
		},
		{
			name:   "failure",
			sig:    "sha256=1234567890abcdef",
			secret: "secret",
			want:   http.StatusForbidden,
		},
		{
			name:   "success",
			sig:    "sha256=dc46983557fea127b43af721467eb9b3fde2338fe3e14f51952aa8478c13d355",
			secret: "secret",
			want:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := listeners.RequestData{
				Headers: http.Header{
					signatureHeader: []string{tt.sig},
				},
				LinkSecrets: map[string]string{
					"webhook_secret": tt.secret,
				},
				RawPayload: []byte("body"),
			}

			if got := CheckSignatureHeader(slog.Default(), r); got != tt.want {
				t.Errorf("checkSignatureHeader() = %d, want %d", got, tt.want)
			}
		})
	}
}
