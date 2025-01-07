package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"strconv"
	"strings"
)

type Middlewares struct {
	AuthService *LogAuthService
}

func NewMiddlewares(authService *LogAuthService) *Middlewares {
	return &Middlewares{
		AuthService: authService,
	}
}

func (m *Middlewares) GrpcAuth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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
			e := catcher.Error("missing connection key", nil, nil)
			e.GinError(c)
			return
		}
		isValid := m.AuthService.IsConnectionKeyValid(connectionKey)
		if !isValid {
			e := catcher.Error("invalid connection key", nil, nil)
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
			e := catcher.Error("failed to read request body", err, nil)
			e.GinError(c)
			return
		}
		sig := c.GetHeader("X-Hub-Signature-256")
		if len(sig) == 0 {
			e := catcher.Error("missing X-Hub-Signature-256 header", nil, nil)
			e.GinError(c)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		key := m.AuthService.GetConnectionKey()
		err = verifySignature(body, key, sig)
		if err != nil {
			e := catcher.Error("failed to verify signature", err, nil)
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

	authKey := md.Get("key")
	authId := md.Get("id")
	connectorType := md.Get("type")
	authConnectionKey := md.Get("connection-key")
	authInternalKey := md.Get("internal-key")

	if len(authKey) > 0 && len(authId) > 0 && len(connectorType) > 0 {
		key := authKey[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return status.Error(codes.PermissionDenied, "id is not valid")
		}
		typ := strings.ToLower(connectorType[0])

		if !m.AuthService.IsKeyValid(key, uint(id), typ) {
			return status.Error(codes.PermissionDenied, "invalid key")
		}
	} else if len(authConnectionKey) > 0 {
		if !isConnectionKeyValid(authConnectionKey[0]) {
			return status.Error(codes.PermissionDenied, "invalid connection key")
		}
	} else if len(authInternalKey) > 0 {
		internalKey := plugins.PluginCfg("com.utmstack", false).Get("internalKey").String()
		if internalKey != authInternalKey[0] {
			return status.Error(codes.PermissionDenied, "internal key does not match")
		}
	} else {
		return status.Error(codes.Unauthenticated, "auth is not provided")
	}

	return nil
}

func verifySignature(payloadBody []byte, secretToken string, signatureHeader string) error {
	if signatureHeader == "" {
		return errors.New("x-hub-signature-256 header is missing")
	}

	mac := hmac.New(sha256.New, []byte(secretToken))
	mac.Write(payloadBody)
	expectedSignature := "sha256=" + fmt.Sprintf("%x", mac.Sum(nil))

	if signatureHeader != expectedSignature {
		return errors.New("request signatures didn't match")
	}

	return nil
}

func isConnectionKeyValid(token string) bool {
	panelKey, e := GetConnectionKey()
	if e != nil {
		return false
	}

	return token == string(panelKey)
}
