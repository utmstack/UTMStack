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

func AgentStreamAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Retrieve the metadata from the context
	routes := config.AgentKeyAuthRoutes()

	if isInRoute(info.FullMethod, routes) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		// Get the authorization token from the metadata
		authToken := md.Get("agent-key")
		authId := md.Get("agent-id")
		if len(authToken) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}
		if len(authId) == 0 {
			return nil, status.Error(codes.Unauthenticated, "agent id is not provided")
		}

		token := authToken[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "agent id is not valid")
		}
		// Replace this with the actual token cache
		tokenCache := agent.Cache
		// Check if the token exists in the token cache
		if _, ok := tokenCache[uint(id)]; !ok || tokenCache[uint(id)] != token {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}
	// Call the handler if the token is valid
	return handler(ctx, req)
}

func InterceptorAgentService(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

func ProcessCommandInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	if info.FullMethod == "/agent.PanelService/ProcessCommand" {
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
