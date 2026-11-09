package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
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

func (m *Middlewares) GrpcAuth(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := m.authFromContext(ctx, si.FullMethod); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (m *Middlewares) GrpcStreamAuth(srv any, ss grpc.ServerStream, si *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := m.authFromContext(ss.Context(), si.FullMethod); err != nil {
		return err
	}
	return handler(srv, ss)
}

func (m *Middlewares) HttpAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		connectionKey := c.GetHeader(proxyAPIKeyHeader)
		if connectionKey == "" {
			e := catcher.Error("cannot authenticate", errors.New("missing connection key"), map[string]any{"process": "plugin_com.utmstack.inputs", "status": http.StatusUnauthorized})
			e.GinError(c)
			return
		}
		isValid := m.AuthService.IsConnectionKeyValid(connectionKey)
		if !isValid {
			e := catcher.Error("cannot authenticate", errors.New("invalid connection key"), map[string]any{"process": "plugin_com.utmstack.inputs", "status": http.StatusUnauthorized})
			e.GinError(c)
			return
		}
		c.Next()
	}
}

// maxWebhookBody caps the unauthenticated webhook body that is buffered before
// the signature is verified, so an anonymous caller cannot drive memory use with
// a single large request.
const maxWebhookBody = 25 << 20 // 25 MiB

func (m *Middlewares) GitHubAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBody)
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
		key := m.AuthService.GetConnectionKey()
		err = verifySignature(body, key, sig)
		if err != nil {
			e := catcher.Error("failed to verify signature", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
			e.GinError(c)
			return
		}

		c.Next()
	}
}

// internalKeyMethods: the ONLY gRPC methods a bare internal-key may invoke.
// The internal key is a deployment-wide machine-to-machine credential; it must
// never authorize the data plane (ProcessLog) — only health. Mirrors the
// InternalKeyRoutes() pattern in agent-manager. (C2 hardening.)
var internalKeyMethods = map[string]struct{}{
	"/grpc.health.v1.Health/Check": {},
}

// internalKeyPermitted reports whether a bare internal-key may invoke the given
// gRPC method. It is the security-relevant decision for the internal-key branch
// and is factored out so it can be tested in isolation without a live
// agent-manager or plugin config.
func internalKeyPermitted(fullMethod string) bool {
	_, ok := internalKeyMethods[fullMethod]
	return ok
}

func (m *Middlewares) authFromContext(ctx context.Context, fullMethod string) error {
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
		if !m.AuthService.IsConnectionKeyValid(authConnectionKey[0]) {
			return status.Error(codes.PermissionDenied, "invalid connection key")
		}
	} else if len(authInternalKey) > 0 {
		// Reject non-allowlisted methods BEFORE touching the key, so a bare
		// internal-key can never reach the data plane (ProcessLog).
		if !internalKeyPermitted(fullMethod) {
			return status.Error(codes.PermissionDenied, "internal-key not permitted for this method")
		}

		internalKey := plugins.PluginCfg("com.utmstack").Get("internalKey").String()
		if subtle.ConstantTimeCompare([]byte(internalKey), []byte(authInternalKey[0])) != 1 {
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

	// Constant-time: a byte-by-byte comparison leaks the expected signature one
	// position at a time to a caller that can retry.
	if !hmac.Equal([]byte(signatureHeader), []byte(expectedSignature)) {
		return errors.New("request signatures didn't match")
	}

	return nil
}
