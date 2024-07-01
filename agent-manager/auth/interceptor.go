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

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func HTTPAuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader("connection-key")
		id := c.GetHeader("id")
		key := c.GetHeader("key")

		if connectionKey == "" && id == "" && key == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "authentication is not provided"})
			return
		} else if connectionKey != "" {
			isValid := validateToken(connectionKey)
			if !isValid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid connection key"})
				return
			}
		} else if id != "" && key != "" {
			idInt, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
				return
			}

			err = checkKeyAuth(key, idInt, c.FullPath(), "http")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth type"})
			return
		}

		c.Next()
	}
}

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
	authConnectionKey := md.Get("connection-key")
	authInternalKey := md.Get("internal-key")

	if len(authKey) > 0 && len(authId) > 0 {
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
		return status.Error(codes.Unauthenticated, fmt.Sprintf("route is not registered for authentication with %s auth type", authType))
	}

	switch authType {
	case "key":
		key := authKey[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return status.Error(codes.Unauthenticated, "id is not valid")
		}

		err = checkKeyAuth(key, id, fullMethod, "grpc")
		if err != nil {
			return err
		}
	case "connection-key":
		if err := authenticateRequest(authConnectionKey[0], "connection-key"); err != nil {
			return err
		}
	case "internal-key":
		if err := authenticateRequest(authInternalKey[0], "internal-key"); err != nil {
			return err
		}
	}
	return nil
}

func checkKeyAuth(token string, id uint64, fullMethod string, proto string) error {
	h := util.GetLogger()
	authCache := getAuthCache(fullMethod, proto)
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
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	return nil
}

type AuthResponse struct {
	Id  uint
	Key string
}

func getAuthCache(method string, proto string) []AuthResponse {
	switch proto {
	case "grpc":
		if strings.Contains(method, "agent.AgentService") {
			return convertMapToAuthResponses(agent.CacheAgent)
		} else if strings.Contains(method, "agent.CollectorService") {
			return convertMapToAuthResponses(agent.CacheCollector)
		} else if strings.Contains(method, "agent.PingService") {
			return append(convertMapToAuthResponses(agent.CacheAgent), convertMapToAuthResponses(agent.CacheCollector)...)
		}
	case "http":
		if strings.Contains(method, "agent") {
			return convertMapToAuthResponses(agent.CacheAgent)
		} else if strings.Contains(method, "collector") {
			return convertMapToAuthResponses(agent.CacheCollector)
		}
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

func authenticateRequest(authValue, authName string) error {
	h := util.GetLogger()

	if authName == "connection-key" && authValue != "" {
		if !validateToken(authValue) {
			h.ErrorF("authentication failed or unable to connect with the panel")
			return status.Error(codes.Unauthenticated, "authentication failed or unable to connect with the panel")
		}
	} else if authName == "internal-key" && authValue != "" {
		internalKey := os.Getenv(config.UTMSharedKeyEnv)
		if authValue != internalKey {
			h.ErrorF("internal key does not match")
			return status.Error(codes.Unauthenticated, "internal key does not match")
		}
	} else {
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
