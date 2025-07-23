package thrippy

import (
	"testing"

	"github.com/rs/zerolog"
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
				Flags: []cli.Flag{&cli.StringFlag{Name: "thrippy-link-provider"}},
			}
			_ = cmd.Set("thrippy-link-provider", tt.flag)
			got, gotOK := LinkID(zerolog.Nop(), cmd, "provider")
			if got != tt.want {
				t.Errorf("LinkID() got = %v, want %v", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("LinkID() OK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
