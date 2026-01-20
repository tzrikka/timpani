// Package thrippy provides common helper functions for
// the [Thrippy gRPC service]. It is meant to facilitate
// code reuse, not to provide a native layer on top of gRPC.
//
// [Thrippy gRPC service]: https://github.com/tzrikka/thrippy-api/blob/main/proto/thrippy/v1/thrippy.proto
package thrippy

import (
	"context"
	"log/slog"
	"time"

	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

const (
	timeout = 3 * time.Second
)

type LinkClient struct {
	LinkID   string
	grpcAddr string
	creds    credentials.TransportCredentials
}

func NewLinkClient(ctx context.Context, id string, cmd *cli.Command) LinkClient {
	return LinkClient{
		LinkID:   id,
		grpcAddr: cmd.String("thrippy-grpc-address"),
		creds:    SecureCreds(ctx, cmd),
	}
}

// LinkCreds returns the saved secrets of the given Thrippy link, or of the receiver's default link
// if no link ID is given. This function does not distinguish between "not found" and other gRPC
// errors. The output must not be cached as it may change at any time, e.g. OAuth access tokens.
func (t *LinkClient) LinkCreds(ctx context.Context, linkID string) (map[string]string, error) {
	if linkID == "" {
		linkID = t.LinkID
	}

	l := activity.GetLogger(ctx)

	conn, err := grpc.NewClient(t.grpcAddr, grpc.WithTransportCredentials(t.creds))
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err), slog.String("grpc_addr", t.grpcAddr))
		return nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err),
			slog.String("link_id", linkID), slog.String("client_method", "GetCredentials"))
		return nil, err
	}

	return resp.GetCredentials(), nil
}

// LinkData returns the template name and saved secrets of the receiver's Thrippy link.
// This function does not distinguish between "not found" and other gRPC errors. The
// output must not be cached as it may change at any time, e.g. OAuth access tokens.
func (t *LinkClient) LinkData(ctx context.Context) (string, map[string]string, error) {
	l := activity.GetLogger(ctx)

	conn, err := grpc.NewClient(t.grpcAddr, grpc.WithTransportCredentials(t.creds))
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err), slog.String("grpc_addr", t.grpcAddr))
		return "", nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Template.
	resp1, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{
		LinkId: proto.String(t.LinkID),
	}.Build())
	if err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err),
			slog.String("link_id", t.LinkID), slog.String("client_method", "GetLink"))
		return "", nil, err
	}

	// Credentials.
	resp2, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(t.LinkID),
	}.Build())
	if err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err),
			slog.String("link_id", t.LinkID), slog.String("client_method", "GetCredentials"))
		return "", nil, err
	}

	return resp1.GetTemplate(), resp2.GetCredentials(), nil
}
