package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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

func (m *Middlewares) GrpcAuth(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := m.authFromContext(ctx); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (m *Middlewares) GrpcStreamAuth(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := m.authFromContext(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}

func (m *Middlewares) HttpAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			e := catcher.Error("cannot authenticate", errors.New("missing api key"), map[string]any{"process": "plugin_com.utmstack.inputs", "status": http.StatusUnauthorized})
			e.GinError(c)
			return
		}
		if !apiKeys.valid(apiKey, c.ClientIP()) {
			e := catcher.Error("cannot authenticate", errors.New("invalid api key"), map[string]any{"process": "plugin_com.utmstack.inputs", "status": http.StatusUnauthorized})
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
			e := catcher.Error("failed to read request body", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
			e.GinError(c)
			return
		}
		sig := c.GetHeader("X-Hub-Signature-256")
		if len(sig) == 0 {
			e := catcher.Error("missing X-Hub-Signature-256 header", nil, map[string]any{"process": "plugin_com.utmstack.inputs"})
			e.GinError(c)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		// The GitHub webhook is configured with the internal key as its HMAC secret.
		key := plugins.PluginCfg("com.utmstack").Get("internalKey").String()
		err = verifySignature(body, key, sig)
		if err != nil {
			e := catcher.Error("failed to verify signature", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
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
	authAPIKey := md.Get(apiKeyHeader)
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
	} else if len(authAPIKey) > 0 {
		if !apiKeys.valid(authAPIKey[0], peerIP(ctx)) {
			return status.Error(codes.PermissionDenied, "invalid api key")
		}
	} else if len(authInternalKey) > 0 {
		internalKey := plugins.PluginCfg("com.utmstack").Get("internalKey").String()
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

// peerIP extracts the client IP of a gRPC call (for the API key IP allowlist).
func peerIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}
	if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
		return host
	}
	return p.Addr.String()
}
