package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestFlags(t *testing.T) {
	if len(flags()) == 0 {
		t.Errorf("flags() should never be nil or empty")
	}
}

func TestConfigDirAndFile(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", d)

	got := configFile()
	want := filepath.Join(d, ConfigDirName, ConfigFileName)
	if got.SourceURI() != want {
		t.Errorf("configFile() = %q, want %q", got.SourceURI(), want)
	}
}

func TestSendHealthzRequest(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "success_200_ok",
			statusCode: http.StatusOK,
		},
		{
			name:       "success_204_no_content",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "error_400_bad_request",
			statusCode: http.StatusBadRequest,
			body:       "invalid request",
			wantErr:    true,
		},
		{
			name:       "error_500_internal_server_error",
			statusCode: http.StatusInternalServerError,
			body:       "server error",
			wantErr:    true,
		},
		{
			name:       "error_503_service_unavailable",
			statusCode: http.StatusServiceUnavailable,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			addr, ok := server.Listener.Addr().(*net.TCPAddr)
			if !ok {
				t.Fatalf("failed to get TCP address from listener")
			}

			gotErr := sendHealthzRequest(t.Context(), addr.Port)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("sendHealthzRequest() = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
