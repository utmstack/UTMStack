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
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	h := util.GetLogger()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		h.ErrorF("metadata is not provided")
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	if isInRoute(info.FullMethod, config.KeyAuthRoutes()) {
		authToken := md.Get("key")
		authId := md.Get("id")
		if len(authToken) == 0 {
			h.ErrorF("authorization token is not provided")
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}
		if len(authId) == 0 {
			h.ErrorF("id is not provided")
			return nil, status.Error(codes.Unauthenticated, "id is not provided")
		}

		token := authToken[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			h.ErrorF("id is not valid")
			return nil, status.Error(codes.Unauthenticated, "id is not valid")
		}

		err = checkKeyAuth(token, id, info.FullMethod)
		if err != nil {
			return nil, err
		}
	} else if isInRoute(info.FullMethod, config.ConnectionKeyRoutes()) {
		if err := authenticateRequest(md, "connection-key"); err != nil {
			return nil, err
		}
	} else if isInRoute(info.FullMethod, config.InternalKeyRoutes()) {
		if err := authenticateRequest(md, "internal-key"); err != nil {
			return nil, err
		}
	}
	return handler(ctx, req)
}

func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	h := util.GetLogger()
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		h.ErrorF("metadata is not provided")
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	if isInRoute(info.FullMethod, config.KeyAuthRoutes()) {
		authToken := md.Get("key")
		authId := md.Get("id")
		if len(authToken) == 0 {
			h.ErrorF("authorization token is not provided")
			return status.Error(codes.Unauthenticated, "authorization token is not provided")
		}
		if len(authId) == 0 {
			h.ErrorF("id is not provided")
			return status.Error(codes.Unauthenticated, "id is not provided")
		}

		token := authToken[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			h.ErrorF("id is not valid")
			return status.Error(codes.Unauthenticated, "id is not valid")
		}

		err = checkKeyAuth(token, id, info.FullMethod)
		if err != nil {
			return err
		}
	} else if isInRoute(info.FullMethod, config.ConnectionKeyRoutes()) {
		if err := authenticateRequest(md, "connection-key"); err != nil {
			return err
		}
	} else if isInRoute(info.FullMethod, config.InternalKeyRoutes()) {
		if err := authenticateRequest(md, "internal-key"); err != nil {
			return err
		}
	}
	return handler(srv, ss)
}

func checkKeyAuth(token string, id uint64, fullMethod string) error {
	h := util.GetLogger()
	authCache := getAuthCache(fullMethod)
	if authCache == nil {
		h.ErrorF("unable to resolve auth cache")
		return status.Error(codes.Unauthenticated, "unable to resolve auth cache")
	}

	found := false
	for _, auth := range authCache {
		if auth.Id == uint(id) && auth.Key == token {
			found = true
			break
		}
	}

	if !found {
		h.ErrorF("invalid token")
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	return nil
}

type AuthResponse struct {
	Id  uint
	Key string
}

func getAuthCache(method string) []AuthResponse {
	if strings.Contains(method, "agent.AgentService") {
		return convertMapToAuthResponses(agent.CacheAgent)
	} else if strings.Contains(method, "agent.CollectorService") {
		return convertMapToAuthResponses(agent.CacheCollector)
	} else if strings.Contains(method, "agent.PingService") {
		return append(convertMapToAuthResponses(agent.CacheAgent), convertMapToAuthResponses(agent.CacheCollector)...)
	}
	return nil
}

func convertMapToAuthResponses(m map[uint]string) []AuthResponse {
	var responses []AuthResponse
	for id, key := range m {
		responses = append(responses, AuthResponse{
			Id:  id,
			Key: key,
		})
	}
	return responses
}

func authenticateRequest(md metadata.MD, authName string) error {
	h := util.GetLogger()
	authHeader := md.Get(authName)

	if len(authHeader) == 0 {
		h.ErrorF("%s must be provided", authName)
		return status.Error(codes.Unauthenticated, fmt.Sprintf("%s must be provided", authName))
	}

	if authName == "connection-key" && authHeader[0] != "" {
		if !validateToken(authHeader[0]) {
			h.ErrorF("unable to connect with the panel to check the connection-key")
			return status.Error(codes.Unauthenticated, "unable to connect with the panel to check the connection-key")
		}
	} else if authName == "internal-key" && authHeader[0] != "" {
		internalKey := os.Getenv(config.UTMSharedKeyEnv)
		if authHeader[0] != internalKey {
			h.ErrorF("internal key does not match")
			return status.Error(codes.Unauthenticated, "internal key does not match")
		}
	} else {
		h.ErrorF("invalid auth name")
		return status.Error(codes.Unauthenticated, "invalid auth name")
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
