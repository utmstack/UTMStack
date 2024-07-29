package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	if err := m.authFromContext(ctx); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (m *Middlewares) GrpcStreamAuth(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := m.authFromContext(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
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

func (m *Middlewares) authFromContext(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "metadata is not provided")
	}

	key, ok := md["key"]
	if !ok || len(key) == 0 {
		return status.Error(codes.Unauthenticated, "authorization key is not provided")
	}

	idStr, ok := md["id"]
	if !ok || len(idStr) == 0 {
		return status.Error(codes.Unauthenticated, "authorization id is not provided")
	}

	connectorType, ok := md["type"]
	if !ok || len(connectorType) == 0 {
		return status.Error(codes.Unauthenticated, "connector type is not provided")
	}

	id, err := strconv.ParseUint(idStr[0], 10, 32)
	if err != nil {
		return status.Error(codes.PermissionDenied, "id is not valid")
	}
	typ := strings.ToLower(connectorType[0])

	if !m.AuthService.IsKeyValid(key[0], uint(id), typ) {
		return status.Error(codes.PermissionDenied, "invalid key")
	}

	return nil
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
