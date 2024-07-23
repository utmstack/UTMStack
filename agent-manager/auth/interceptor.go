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
			isValid := isConnectionKeyValid(connectionKey)
			if !isValid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid connection key"})
				return
			}
		} else if id != "" && key != "" {
			idInt, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
				return
			}

			if _, _, isValid := agent.ValidateKeyPairFromCache(key, uint(idInt)); !isValid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
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
		return status.Error(codes.PermissionDenied, fmt.Sprintf("route is not registered for authentication with %s auth type", authType))
	}

	switch authType {
	case "key":
		key := authKey[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return status.Error(codes.PermissionDenied, "id is not valid")
		}

		if _, _, isValid := agent.ValidateKeyPairFromCache(key, uint(id)); !isValid {
			return status.Error(codes.PermissionDenied, "invalid key")
		}
	case "connection-key":
		if !isConnectionKeyValid(authConnectionKey[0]) {
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
	internalKey := os.Getenv(config.UTMSharedKeyEnv)
	return token == internalKey
}

func isConnectionKeyValid(token string) bool {
	url := fmt.Sprintf(config.PanelConnectionKeyUrl, os.Getenv(config.UTMHostEnv))
	requestBody := strings.NewReader(token)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
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
