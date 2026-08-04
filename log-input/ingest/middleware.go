package ingest

import (
	"context"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/utmstack/UTMStack/log-input/auth"
)

const apiKeyHeader = "Utm-Api-Key"

type middlewares struct {
	auth *auth.Service
}

func newMiddlewares(a *auth.Service) *middlewares {
	return &middlewares{auth: a}
}

func (m *middlewares) unary(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := m.check(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// stream authenticates once, when the stream opens, so this is not on the
// per-log path.
func (m *middlewares) stream(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := m.check(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}

func (m *middlewares) check(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "metadata is not provided")
	}

	authKey := md.Get("key")
	authID := md.Get("id")
	connectorType := md.Get("type")
	apiKey := md.Get(apiKeyHeader)
	internalKey := md.Get("internal-key")

	switch {
	case len(authKey) > 0 && len(authID) > 0 && len(connectorType) > 0:
		id, err := strconv.ParseUint(authID[0], 10, 64)
		if err != nil {
			return status.Error(codes.PermissionDenied, "id is not valid")
		}
		if !m.auth.ConnectorValid(ctx, authKey[0], id, connectorType[0]) {
			return status.Error(codes.PermissionDenied, "invalid key")
		}

	case len(apiKey) > 0:
		if !m.auth.APIKeyValid(ctx, apiKey[0], peerIP(ctx)) {
			return status.Error(codes.PermissionDenied, "invalid api key")
		}

	case len(internalKey) > 0:
		if !m.auth.InternalKeyValid(internalKey[0]) {
			return status.Error(codes.PermissionDenied, "internal key does not match")
		}

	default:
		return status.Error(codes.Unauthenticated, "auth is not provided")
	}

	return nil
}

func peerIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}
	if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
		return host
	}
	return p.Addr.String()
}
