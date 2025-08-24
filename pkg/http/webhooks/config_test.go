package webhooks

import (
	"testing"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "-65536",
			port:    -65536,
			wantErr: true,
		},
		{
			name:    "-1",
			port:    -1,
			wantErr: true,
		},
		{
			name: "0",
			port: 0,
		},
		{
			name: "1",
			port: 1,
		},
		{
			name: "65535",
			port: 65535,
		},
		{
			name:    "65536",
			port:    65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePort(tt.port); (err != nil) != tt.wantErr {
				t.Errorf("validatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
