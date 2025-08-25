// Package thrippy provides common helper functions for
// the [Thrippy gRPC service]. It is meant to facilitate
// code reuse, not to provide a native layer on top of gRPC.
//
// [Thrippy gRPC service]: https://github.com/tzrikka/thrippy-api/blob/main/proto/thrippy/v1/thrippy.proto
package thrippy

import (
	"context"
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

func NewLinkClient(id string, cmd *cli.Command) LinkClient {
	return LinkClient{
		LinkID:   id,
		grpcAddr: cmd.String("thrippy-grpc-address"),
		creds:    SecureCreds(cmd),
	}
}

// LinkCreds returns the saved secrets of the receiver's Thrippy link. This
// function does not distinguish between "not found" and other gRPC errors. The
// output must not be cached as it may change at any time, e.g. OAuth access tokens.
func (t *LinkClient) LinkCreds(ctx context.Context) (map[string]string, error) {
	l := activity.GetLogger(ctx)

	conn, err := grpc.NewClient(t.grpcAddr, grpc.WithTransportCredentials(t.creds))
	if err != nil {
		l.Error("failed to create gRPC client connection", "error", err, "grpc_addr", t.grpcAddr)
		return nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(t.LinkID),
	}.Build())
	if err != nil {
		l.Error("Thrippy GetCredentials error", "error", err, "link_id", t.LinkID)
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
		l.Error("failed to create gRPC client connection", "error", err, "grpc_addr", t.grpcAddr)
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
		l.Error("Thrippy GetLink error", "error", err, "link_id", t.LinkID)
		return "", nil, err
	}

	// Credentials.
	resp2, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(t.LinkID),
	}.Build())
	if err != nil {
		l.Error("Thrippy GetCredentials error", "error", err, "link_id", t.LinkID)
		return "", nil, err
	}

	return resp1.GetTemplate(), resp2.GetCredentials(), nil
}
