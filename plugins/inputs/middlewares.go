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
	go_sdk "github.com/threatwinds/go-sdk"
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
			e := go_sdk.Logger().ErrorF("connection key not provided")
			e.GinError(c)
			return
		}
		isValid := m.AuthService.IsConnectionKeyValid(connectionKey)
		if !isValid {
			e := go_sdk.Logger().ErrorF("invalid connection key")
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
			e := go_sdk.Logger().ErrorF("error reading request body: %v", err)
			e.GinError(c)
			return
		}
		sig := c.GetHeader("X-Hub-Signature-256")
		if len(sig) == 0 {
			e := go_sdk.Logger().ErrorF("missing X-Hub-Signature-256")
			e.GinError(c)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		key := m.AuthService.GetConnectionKey()
		err = verifySignature(body, key, sig)
		if err != nil {
			e := go_sdk.Logger().ErrorF(err.Error())
			e.GinError(c)
			return
		}

		c.Next()
	}
}

func (m *Middlewares) authFromContext(ctx context.Context) error {
	cnf, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		return status.Error(codes.Internal, "failed to get config")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "metadata is not provided")
	}

	var authType string
	authKey := md.Get("key")
	authId := md.Get("id")
	connectorType := md.Get("type")
	authConnectionKey := md.Get("connection-key")
	authInternalKey := md.Get("internal-key")

	if len(authKey) > 0 && len(authId) > 0 && len(connectorType) > 0 {
		authType = "key"
	} else if len(authConnectionKey) > 0 {
		authType = "connection-key"
	} else if len(authInternalKey) > 0 {
		authType = "internal-key"
	} else {
		return status.Error(codes.Unauthenticated, "auth is not provided")
	}

	switch authType {
	case "key":
		key := authKey[0]
		id, err := strconv.ParseUint(authId[0], 10, 32)
		if err != nil {
			return status.Error(codes.PermissionDenied, "id is not valid")
		}
		typ := strings.ToLower(connectorType[0])

		if !m.AuthService.IsKeyValid(key, uint(id), typ) {
			return status.Error(codes.PermissionDenied, "invalid key")
		}
	case "connection-key":
		if !isConnectionKeyValid(authConnectionKey[0]) {
			return status.Error(codes.PermissionDenied, "invalid connection key")
		}
	case "internal-key":
		if cnf.InternalKey != authInternalKey[0] {
			return status.Error(codes.PermissionDenied, "internal key does not match")
		}
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
