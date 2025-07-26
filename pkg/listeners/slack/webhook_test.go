package slack

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"

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
				Headers: http.Header{contentTypeHeader: []string{tt.ct}},
			}

			if got := checkContentTypeHeader(zerolog.Nop(), r); got != tt.want {
				t.Errorf("checkContentTypeHeader() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckTimestampHeader(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name string
		ts   string
		want int
	}{
		{
			name: "none",
			want: http.StatusBadRequest,
		},
		{
			name: "invalid_timestamp",
			ts:   "kaboom",
			want: http.StatusBadRequest,
		},
		{
			name: "fresh",
			ts:   strconv.FormatInt(now-10, 10),
			want: http.StatusOK,
		},
		{
			name: "stale",
			ts:   strconv.FormatInt(now-360, 10),
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := listeners.RequestData{
				Headers: http.Header{
					timestampHeader: []string{tt.ts},
				},
			}

			if got := checkTimestampHeader(zerolog.Nop(), r); got != tt.want {
				t.Errorf("checkTimestampHeader() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckSignatureHeader(t *testing.T) {
	tests := []struct {
		name   string
		sig    string
		secret string
		ts     string
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
			sig:    "v0=1234567890abcdef",
			secret: "secret",
			ts:     "100000",
			want:   http.StatusForbidden,
		},
		{
			name:   "success",
			sig:    "v0=805ceef08cf066824eb49058aabfcd59c33759a201e9405cbdba329920e68045",
			secret: "secret",
			ts:     "100000",
			want:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := listeners.RequestData{
				Headers: http.Header{
					signatureHeader: []string{tt.sig},
					timestampHeader: []string{tt.ts},
				},
				LinkSecrets: map[string]string{
					"signing_secret": tt.secret,
				},
				RawPayload: []byte("body"),
			}

			if got := checkSignatureHeader(zerolog.Nop(), r); got != tt.want {
				t.Errorf("checkSignatureHeader() = %d, want %d", got, tt.want)
			}
		})
	}
}
