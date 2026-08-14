package ingest

import (
	"context"
	"errors"
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

var errUnauthenticated = errors.New("auth is not provided")

var errPermissionDenied = errors.New("invalid credential")

type tenantResolver interface {
	APIKeyTenant(ctx context.Context, apiKey, clientIP string) (string, bool)
	ConnectorTenant(ctx context.Context, key string, id uint64, typ string) (string, bool)
}

type grpcResolver interface {
	tenantResolver
	InternalKeyValid(key string) bool
}

type middlewares struct {
	auth grpcResolver
}

func newMiddlewares(a *auth.Service) *middlewares {
	return &middlewares{auth: a}
}

type Credentials struct {
	APIKey   string
	ConnKey  string
	ConnID   uint64
	ConnType string
	ClientIP string
}

func resolveAuth(ctx context.Context, resolver tenantResolver, c Credentials) (tenant string, err error) {
	switch {
	case c.APIKey != "":
		t, ok := resolver.APIKeyTenant(ctx, c.APIKey, c.ClientIP)
		if !ok {
			return "", errPermissionDenied
		}
		return t, nil

	case c.ConnKey != "" && c.ConnID != 0 && c.ConnType != "":
		t, ok := resolver.ConnectorTenant(ctx, c.ConnKey, c.ConnID, c.ConnType)
		if !ok {
			return "", errPermissionDenied
		}
		return t, nil

	default:
		return "", errUnauthenticated
	}
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

	if keys := md.Get("internal-key"); len(keys) > 0 {
		if !m.auth.InternalKeyValid(keys[0]) {
			return nil, status.Error(codes.PermissionDenied, "internal key does not match")
		}
		return ctx, nil
	}

	creds := Credentials{ClientIP: peerIP(ctx)}
	if v := md.Get("key"); len(v) > 0 {
		creds.ConnKey = v[0]
	}
	if v := md.Get("id"); len(v) > 0 {
		id, err := strconv.ParseUint(v[0], 10, 64)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "id is not valid")
		}
		creds.ConnID = id
	}
	if v := md.Get("type"); len(v) > 0 {
		creds.ConnType = v[0]
	}
	if v := md.Get(apiKeyHeader); len(v) > 0 {
		creds.APIKey = v[0]
	}

	tenant, err := resolveAuth(ctx, m.auth, creds)
	if err != nil {
		switch {
		case errors.Is(err, errUnauthenticated):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
	}
	return WithTenant(ctx, tenant), nil
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
