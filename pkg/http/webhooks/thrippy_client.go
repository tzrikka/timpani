package webhooks

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

// linkData returns the template name and saved secrets of the given Thrippy link.
// This function reports gRPC errors, but if the link is not found it returns nothing.
// The output must not be cached as it may change at any time, e.g. OAuth access tokens.
func (s *httpServer) linkData(ctx context.Context, linkID string) (string, map[string]string, error) {
	l := zerolog.Ctx(ctx)

	conn, err := grpc.NewClient(s.thrippyGRPCAddr, grpc.WithTransportCredentials(s.thrippyCreds))
	if err != nil {
		l.Error().Stack().Err(err).Send()
		return "", nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Template.
	resp1, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		if status.Code(err) != codes.NotFound {
			l.Error().Stack().Err(err).Send()
			return "", nil, err
		}
		return "", nil, nil
	}

	// Credentials.
	resp2, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		l.Error().Stack().Err(err).Send()
		return "", nil, err
	}

	return resp1.GetTemplate(), resp2.GetCredentials(), nil
}

func checkLinkData(l zerolog.Logger, template string, secrets map[string]string, err error) int {
	if err != nil {
		l.Warn().Err(err).Msg("failed to get link secrets from Thrippy over gRPC")
		return http.StatusInternalServerError
	}

	if template == "" && secrets == nil {
		l.Warn().Msg("bad request: link not found")
		return http.StatusNotFound
	}

	if template != "" && secrets == nil {
		l.Warn().Msg("bad request: link not initialized")
		return http.StatusNotFound
	}

	return http.StatusOK
}
