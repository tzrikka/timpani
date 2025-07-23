package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOrCachedClient(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "upgrade")
		w.Header().Set("Sec-WebSocket-Accept", "BACScCJPNqyz+UBoqMH89VmURoA=")
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))
	defer s.Close()

	url := func(_ context.Context) (string, error) {
		return s.URL, nil
	}

	tests := []struct {
		name    string
		id      string
		wantLen int
	}{
		{
			name:    "store_first_client",
			id:      "1",
			wantLen: 1,
		},
		{
			name:    "store_second_client",
			id:      "2",
			wantLen: 2,
		},
		{
			name:    "load_first_client",
			id:      "1",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewOrCachedClient(t.Context(), url, tt.id, withTestNonceGen()); err != nil {
				t.Fatalf("NewOrCachedClient() error = %v", err)
			}

			if l := lenClients(); l != tt.wantLen {
				t.Fatalf("len(clients) == %d, want %d", l, tt.wantLen)
			}
		})
	}
}

func lenClients() int {
	count := 0
	clients.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func TestHash(t *testing.T) {
	h1, h2, h3 := hash("1"), hash("2"), hash("1")
	if h1 == h2 {
		t.Errorf("hash() isn't unique: %q == %q", h1, h2)
	}
	if h1 != h3 {
		t.Errorf("hash() isn't stable: %q != %q", h1, h2)
	}
}
