package webhooks

import (
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

type server struct {
	thrippypb.UnimplementedThrippyServiceServer

	linkResp  *thrippypb.GetLinkResponse
	credsResp *thrippypb.GetCredentialsResponse
	err       error
}

func (s *server) GetLink(_ context.Context, _ *thrippypb.GetLinkRequest) (*thrippypb.GetLinkResponse, error) {
	return s.linkResp, s.err
}

func (s *server) GetCredentials(_ context.Context, _ *thrippypb.GetCredentialsRequest) (*thrippypb.GetCredentialsResponse, error) {
	return s.credsResp, s.err
}

func TestHTTPServerLinkData(t *testing.T) {
	tests := []struct {
		name         string
		linkResp     *thrippypb.GetLinkResponse
		credsResp    *thrippypb.GetCredentialsResponse
		respErr      error
		wantTemplate string
		wantSecrets  map[string]string
		wantErr      bool
	}{
		{
			name: "nil",
		},
		{
			name:    "grpc_error",
			respErr: errors.New("error"),
			wantErr: true,
		},
		{
			name:    "link_not_found",
			respErr: status.Error(codes.NotFound, "link not found"),
		},
		{
			name: "existing_link_without_secrets",
			linkResp: thrippypb.GetLinkResponse_builder{
				Template: proto.String("template"),
			}.Build(),
			credsResp:    thrippypb.GetCredentialsResponse_builder{}.Build(),
			wantTemplate: "template",
		},
		{
			name: "happy_path",
			linkResp: thrippypb.GetLinkResponse_builder{
				Template: proto.String("template"),
			}.Build(),
			credsResp: thrippypb.GetCredentialsResponse_builder{
				Credentials: map[string]string{"aaa": "111", "bbb": "222"},
			}.Build(),
			wantTemplate: "template",
			wantSecrets:  map[string]string{"aaa": "111", "bbb": "222"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := net.ListenConfig{}
			lis, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			gs := grpc.NewServer()
			thrippypb.RegisterThrippyServiceServer(gs, &server{
				linkResp:  tt.linkResp,
				credsResp: tt.credsResp,
				err:       tt.respErr,
			})
			go func() {
				_ = gs.Serve(lis)
			}()

			hs := &httpServer{
				thrippyGRPCAddr: lis.Addr().String(),
				thrippyCreds:    insecure.NewCredentials(),
			}

			template, secrets, err := hs.linkData(t.Context(), "link ID")
			if (err != nil) != tt.wantErr {
				t.Errorf("linkData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if template != tt.wantTemplate {
				t.Errorf("linkData() template = %q, want %q", template, tt.wantTemplate)
			}
			if !reflect.DeepEqual(secrets, tt.wantSecrets) {
				t.Errorf("linkData() secrets = %v, want %v", secrets, tt.wantSecrets)
			}
		})
	}
}

func TestCheckLinkDataForWebhook(t *testing.T) {
	tests := []struct {
		name     string
		template string
		secrets  map[string]string
		err      error
		want     int
	}{
		{
			name: "internal_error",
			err:  errors.New("some error"),
			want: http.StatusInternalServerError,
		},
		{
			name: "not_found",
			want: http.StatusNotFound,
		},
		{
			name:     "no_secrets",
			template: "name",
			want:     http.StatusNotFound,
		},
		{
			name:     "happy_path",
			template: "name",
			secrets: map[string]string{
				"foo": "bar",
			},
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkLinkDataForWebhook(zerolog.Nop(), tt.template, tt.secrets, tt.err); got != tt.want {
				t.Errorf("checkLinkDataForWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckLinkDataForConn(t *testing.T) {
	tests := []struct {
		name     string
		template string
		secrets  map[string]string
		err      error
		wantErr  bool
	}{
		{
			name:    "internal_error",
			err:     errors.New("some error"),
			wantErr: true,
		},
		{
			name:    "not_found",
			wantErr: true,
		},
		{
			name:     "no_secrets",
			template: "name",
			wantErr:  true,
		},
		{
			name:     "happy_path",
			template: "name",
			secrets: map[string]string{
				"foo": "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkLinkDataForConn(zerolog.Nop(), tt.template, tt.secrets, tt.err); (got != nil) != tt.wantErr {
				t.Errorf("checkLinkDataForConn() err = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
