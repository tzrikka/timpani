package thrippy

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestLinkID(t *testing.T) {
	tests := []struct {
		name   string
		flag   string
		want   string
		wantOK bool
	}{
		{
			name: "no_value",
		},
		{
			name: "invalid_value",
			flag: "1",
		},
		{
			name:   "happy_path",
			flag:   "jPQqh6z3mahiw5xFtnyESK",
			want:   "jPQqh6z3mahiw5xFtnyESK",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cli.Command{
				Flags: []cli.Flag{&cli.StringFlag{Name: "thrippy-link-service"}},
			}
			_ = cmd.Set("thrippy-link-service", tt.flag)
			got, gotOK := LinkID(cmd, "service")
			if got != tt.want {
				t.Errorf("LinkID() got = %v, want %v", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("LinkID() OK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

func TestValidateOptionalUUID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: false,
		},
		{
			name:    "valid_uuid",
			id:      "jPQqh6z3mahiw5xFtnyESK",
			wantErr: false,
		},
		{
			name:    "invalid_uuid",
			id:      "invalid-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOptionalUUID(tt.id); (err != nil) != tt.wantErr {
				t.Errorf("validateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
