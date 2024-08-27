package agent

import (
	"context"
	_ "errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	err := authHeaders(md, info.FullMethod)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	err := authHeaders(md, info.FullMethod)
	if err != nil {
		return err
	}

	return handler(srv, ss)
}

func authHeaders(md metadata.MD, fullMethod string) error {
	var authType string
	var routes []string
	authKey := md.Get("key")
	authId := md.Get("id")
	connectorType := md.Get("type")
	authConnectionKey := md.Get("connection-key")
	authInternalKey := md.Get("internal-key")

	if len(authKey) > 0 && len(authId) > 0 && len(connectorType) > 0 {
		authType = "key"
		routes = config.KeyAuthRoutes()
	} else if len(authConnectionKey) > 0 {
		authType = "connection-key"
		routes = config.ConnectionKeyRoutes()
	} else if len(authInternalKey) > 0 {
		authType = "internal-key"
		routes = config.InternalKeyRoutes()
	} else {
		return status.Error(codes.Unauthenticated, "auth is not provided")
	}

	if !isInRoute(fullMethod, routes) {
		return status.Error(codes.PermissionDenied, fmt.Sprintf("route is not registered for authentication with %s auth type", authType))
	}

	switch authType {
	case "key":
		key := authKey[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return status.Error(codes.PermissionDenied, "id is not valid")
		}
		typ := strings.ToLower(connectorType[0])
		switch typ {
		case "agent":
			if _, isValid := utils.IsKeyPairValid(key, uint(id), AgentServ.CacheAgentKey); !isValid {
				return status.Error(codes.PermissionDenied, "invalid key")
			}
		case "collector":
			if _, isValid := utils.IsKeyPairValid(key, uint(id), CollectorServ.CacheCollectorKey); !isValid {
				return status.Error(codes.PermissionDenied, "invalid key")
			}
		default:
			return status.Error(codes.PermissionDenied, "invalid type")
		}
	case "connection-key":
		if !utils.IsConnectionKeyValid(fmt.Sprintf(config.PanelConnectionKeyUrl, config.GetUTMHost()), authConnectionKey[0]) {
			return status.Error(codes.PermissionDenied, "invalid connection key")
		}
	case "internal-key":
		if !isInternalKeyValid(authInternalKey[0]) {
			return status.Error(codes.PermissionDenied, "internal key does not match")
		}
	}
	return nil
}

func isInternalKeyValid(token string) bool {
	return token == config.GetInternalKey()
}

func isInRoute(route string, list []string) bool {
	for _, r := range list {
		if r == route {
			return true
		}
	}
	return false
}
