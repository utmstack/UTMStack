package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/helpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Middlewares struct {
	AuthService *LogAuthService
}

func NewMiddlewares(authService *LogAuthService) *Middlewares {
	return &Middlewares{
		AuthService: authService,
	}
}

func (m *Middlewares) GrpcAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	// Get the authorization token from the metadata
	authToken := md.Get("key")
	if len(authToken) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	key := authToken[0]

	if !m.AuthService.IsKeyValid(key) {
		return nil, status.Error(codes.Unauthenticated, "invalid key")
	}

	return handler(ctx, req)
}

func (m *Middlewares) HttpAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader(proxyAPIKeyHeader)
		if connectionKey == "" {
			e := helpers.Logger().ErrorF("connection key not provided")
			e.GinError(c)
			return
		}
		isValid := m.AuthService.IsConnectionKeyValid(connectionKey)
		if !isValid {
			e := helpers.Logger().ErrorF("invalid connection key")
			e.GinError(c)
			return
		}
		c.Next()
	}
}

func (m *Middlewares) GitHubAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			e := helpers.Logger().ErrorF("error reading request body: %v", err)
			e.GinError(c)
			return
		}
		sig := c.GetHeader("X-Hub-Signature-256")
		if len(sig) == 0 {
			e := helpers.Logger().ErrorF("missing X-Hub-Signature-256")
			e.GinError(c)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		key := m.AuthService.GetConnectionKey()
		err = verifySignature(body, key, sig)
		if err != nil {
			e := helpers.Logger().ErrorF(err.Error())
			e.GinError(c)
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
