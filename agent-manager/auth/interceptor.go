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
	// Check if the current method is RegisterAgent or DeleteAgent
	routes := config.ConnectionKeyRoutes()
	if isInRoute(info.FullMethod, routes) {
		// Get the metadata from the incoming context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		// Get the authorization token from metadata
		authHeader := md.Get("connection-key")
		internalKeyHeader := md.Get("internal-key")
		if len(authHeader) == 0 && len(internalKeyHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization key must be provided")
		}
		if authHeader != nil && authHeader[0] != "" {
			if !validateToken(authHeader[0]) {
				return nil, status.Error(codes.Unauthenticated, "unable to connect with the panel to check the connection key")
			}
		} else {
			// check the internal key
			internalKey := os.Getenv(config.UTMSharedKeyEnv)
			if internalKeyHeader == nil || internalKeyHeader[0] != internalKey {
				return nil, status.Error(codes.Unauthenticated, "internal key does not match")
			}
		}
	}
	// If the token is valid or the internal key, call the handler to continue processing the request
	return handler(ctx, req)
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
