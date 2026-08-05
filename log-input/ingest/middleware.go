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
	ctx, err := m.check(ctx)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (m *middlewares) stream(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx, err := m.check(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, &authedStream{ServerStream: ss, ctx: ctx})
}

type authedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authedStream) Context() context.Context { return s.ctx }

func (m *middlewares) check(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "metadata is not provided")
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
			return nil, status.Error(codes.PermissionDenied, "id is not valid")
		}
		tenant, ok := m.auth.ConnectorTenant(ctx, authKey[0], id, connectorType[0])
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "invalid key")
		}
		return WithTenant(ctx, tenant), nil

	case len(apiKey) > 0:
		tenant, ok := m.auth.APIKeyTenant(ctx, apiKey[0], peerIP(ctx))
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "invalid api key")
		}
		return WithTenant(ctx, tenant), nil

	case len(internalKey) > 0:
		if !m.auth.InternalKeyValid(internalKey[0]) {
			return nil, status.Error(codes.PermissionDenied, "internal key does not match")
		}
		return ctx, nil

	default:
		return nil, status.Error(codes.Unauthenticated, "auth is not provided")
	}
}

type tenantKey struct{}

func WithTenant(ctx context.Context, tenant string) context.Context {
	if tenant == "" {
		return ctx
	}
	return context.WithValue(ctx, tenantKey{}, tenant)
}

func tenantFrom(ctx context.Context) string {
	t, _ := ctx.Value(tenantKey{}).(string)
	return t
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
