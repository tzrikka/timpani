package webhooks

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/timpani/internal/logger"
)

// linkData returns the template name and saved secrets of the given Thrippy link.
// This function reports gRPC errors, but if the link is not found it returns nothing.
// The output must not be cached as it may change at any time, e.g. OAuth access tokens.
func (s *HTTPServer) linkData(ctx context.Context, linkID string) (string, map[string]string, error) {
	l := logger.FromContext(ctx).With(slog.String("link_id", linkID))

	conn, err := grpc.NewClient(s.thrippyGRPCAddr, grpc.WithTransportCredentials(s.thrippyCreds))
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err))
		return "", nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	// Template.
	resp1, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		if status.Code(err) != codes.NotFound {
			l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "GetLink"))
			return "", nil, err
		}
		return "", nil, nil
	}

	// Credentials.
	resp2, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "GetCredentials"))
		return "", nil, err
	}

	return resp1.GetTemplate(), resp2.GetCredentials(), nil
}

func checkLinkDataForWebhook(l *slog.Logger, template string, secrets map[string]string, err error) int {
	if err != nil {
		l.Warn("failed to get link secrets from Thrippy over gRPC", slog.Any("error", err))
		return http.StatusInternalServerError
	}

	if template == "" && secrets == nil {
		l.Warn("bad request: link not found")
		return http.StatusNotFound
	}

	if template != "" && secrets == nil {
		l.Warn("bad request: link not initialized")
		return http.StatusNotFound
	}

	return http.StatusOK
}

func checkLinkDataForConn(l *slog.Logger, template string, secrets map[string]string, err error) error {
	if err != nil {
		l.Warn("failed to get link secrets from Thrippy over gRPC", slog.Any("error", err))
		return err
	}

	if template == "" && secrets == nil {
		l.Error("link ID not found in Thrippy")
		return errors.New("configured link ID not found in Thrippy")
	}

	if template != "" && secrets == nil {
		l.Error("Thrippy link not initialized", slog.String("template", template))
		return errors.New("link not initialized in Thrippy")
	}

	return nil
}
