package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LogAuthInterceptor struct {
	Mutex       *sync.Mutex
	AuthService *logservice.LogAuthService
}

func NewLogAuthInterceptor(authService *logservice.LogAuthService) *LogAuthInterceptor {
	return &LogAuthInterceptor{
		Mutex:       &sync.Mutex{},
		AuthService: authService,
	}
}

func (interceptor *LogAuthInterceptor) GrpcAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	// Get the authorization token from the metadata
	authToken := md.Get("agent-key")
	if len(authToken) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	key := authToken[0]
	if !interceptor.AuthService.IsAgentKeyValid(key) {
		return nil, status.Error(codes.Unauthenticated, "invalid agent key")
	}
	return handler(ctx, req)
}
func (interceptor *LogAuthInterceptor) GrpcRecoverInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Handle the panic here
			log.Printf("Panic occurred: %v", r)
			err = status.Errorf(codes.Internal, "Internal server error")
		}
	}()

	// Call the gRPC handler
	return handler(ctx, req)
}

func (interceptor *LogAuthInterceptor) HTTPAuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader(config.ProxyAPIKeyHeader)
		if connectionKey == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Connection key not provided"})
			return
		}
		isValid := interceptor.AuthService.IsConnectionKeyValid(connectionKey)
		if !isValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Connection Key"})
			return
		}
		c.Next()
	}
}
func (interceptor *LogAuthInterceptor) HTTPGitHubAuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, "error reading request body")
			return
		}
		sig := c.GetHeader("X-Hub-Signature-256")
		if len(sig) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Missing X-Hub-Signature-256")
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		key := interceptor.AuthService.GetConnectionKey()
		err = verifySignature(body, key, sig)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid signature")
			return
		}

		c.Next()
	}
}

func verifySignature(payloadBody []byte, secretToken string, signatureHeader string) error {
	if signatureHeader == "" {
		return errors.New("x-hub-signature-256 header is missing")
	}

	// Calculate the expected HMAC
	mac := hmac.New(sha256.New, []byte(secretToken))
	mac.Write(payloadBody)
	expectedSignature := "sha256=" + fmt.Sprintf("%x", mac.Sum(nil))

	// Compare the calculated HMAC to the received signature
	if signatureHeader != expectedSignature {
		return errors.New("request signatures didn't match")
	}

	return nil
}
