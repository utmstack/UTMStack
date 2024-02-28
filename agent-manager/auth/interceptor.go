package auth

import (
	"context"
	"crypto/tls"
	_ "errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func StreamAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Retrieve the metadata from the context
	routes := config.AgentKeyAuthRoutes()
	route := info.FullMethod
	if isInRoute(route, routes) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		// Get the authorization key from the metadata
		authToken := md.Get("key")
		authId := md.Get("id")
		if len(authToken) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}
		if len(authId) == 0 {
			return nil, status.Error(codes.Unauthenticated, "agent id is not provided")
		}

		token := authToken[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "id is not valid")
		}
		// Replace this with the actual token cache
		authCache := getAuthCache(route)
		if authCache == nil {
			return nil, status.Error(codes.Unauthenticated, "unable to resolve auth cache")
		}
		// Check if the token exists in the token cache
		if _, ok := authCache[uint(id)]; !ok || authCache[uint(id)] != token {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}
	// Call the handler if the token is valid
	return handler(ctx, req)
}

func getAuthCache(method string) map[uint]string {
	if strings.Contains(method, "agent.AgentService") {
		return agent.CacheAgent
	} else if strings.Contains(method, "agent.CollectorService") {
		return agent.CacheCollector
	}
	return nil
}

func ConnectionKeyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if isInRoute(info.FullMethod, config.ConnectionKeyRoutes()) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		if err := authenticateRequest(md); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}

func PanelInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	if info.FullMethod == "/agent.PanelService/ProcessCommand" || info.FullMethod == "/agent.CollectorConfigurationService/CollectorConfigStream" {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		if err := authenticateRequest(md); err != nil {
			return err
		}
	}

	return handler(srv, ss)
}

func authenticateRequest(md metadata.MD) error {
	authHeader := md.Get("connection-key")
	internalKeyHeader := md.Get("internal-key")

	if len(authHeader) == 0 && len(internalKeyHeader) == 0 {
		return status.Error(codes.Unauthenticated, "authorization key must be provided")
	}

	if len(authHeader) > 0 && authHeader[0] != "" {
		if !validateToken(authHeader[0]) {
			return status.Error(codes.Unauthenticated, "unable to connect with the panel to check the connection key")
		}
	} else {
		internalKey := os.Getenv(config.UTMSharedKeyEnv)
		if internalKeyHeader == nil || len(internalKeyHeader) == 0 || internalKeyHeader[0] != internalKey {
			return status.Error(codes.Unauthenticated, "internal key does not match")
		}
	}

	return nil
}

func validateToken(token string) bool {
	url := fmt.Sprintf(config.PanelConnectionKeyUrl, os.Getenv(config.UTMHostEnv))
	requestBody := strings.NewReader(token)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	// Use the custom transport
	client := &http.Client{Transport: transport}
	resp, err := client.Post(url, "application/json", requestBody)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func isInRoute(route string, list []string) bool {
	for _, r := range list {
		if r == route {
			return true
		}
	}
	return false
}
