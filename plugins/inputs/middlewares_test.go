package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	sdkplugins "github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// testInternalKey is the deployment-wide internal key the seeded plugin config
// carries. It is used only to exercise the allowlist gate on the internal-key
// branch; the deny cases are rejected before the value is ever compared.
const testInternalKey = "test-internal-key-0123456789abcdef"

// TestMain redirects the go-sdk plugin WorkDir to a temp directory that seeds a
// `com.utmstack` config carrying a known internalKey. This lets the internal-key
// branch of authFromContext resolve its key from real plugin config (via
// plugins.PluginCfg) WITHOUT a live agent-manager or the rest of the deployment.
//
// plugins.WorkDir is an exported var read at call time (config.go loadCfg), so
// reassigning it before the first PluginCfg invocation is enough to point the
// config loader at the seeded file.
func TestMain(m *testing.M) {
	workdir, err := os.MkdirTemp("", "inputs-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(workdir)

	if err := os.MkdirAll(filepath.Join(workdir, "pipeline"), 0o755); err != nil {
		panic(err)
	}

	cfg := "plugins:\n  com.utmstack:\n    internalKey: " + testInternalKey + "\n"
	if err := os.WriteFile(filepath.Join(workdir, "pipeline", "system_plugins.yaml"), []byte(cfg), 0o600); err != nil {
		panic(err)
	}

	sdkplugins.WorkDir = workdir

	os.Exit(m.Run())
}

// testAuthService returns a LogAuthService with empty caches and a known
// connection key, enough for the key/id/type and connection-key branches of
// authFromContext. No network, no gRPC.
func testAuthService() *LogAuthService {
	return &LogAuthService{
		Mutex:              &sync.Mutex{},
		CollectorKeyCache:  map[uint]string{42: "collector-42-key"},
		AgentKeyCache:      map[uint]string{7: "agent-7-key"},
		ConnectionKeyCache: "test-connection-key",
	}
}

func internalKeyCtx() context.Context {
	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("internal-key", testInternalKey))
}

func keyIDTypeCtx(id string, key string, typ string) context.Context {
	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("key", key, "id", id, "type", typ))
}

func connectionKeyCtx(key string) context.Context {
	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("connection-key", key))
}

// TestInternalKeyPermitted pins the allowlist decision in isolation: only the
// health full-method name is permitted; the data-plane method and any unknown
// method are not.
func TestInternalKeyPermitted(t *testing.T) {
	cases := []struct {
		name      string
		method    string
		permitted bool
	}{
		{"health check permitted", "/grpc.health.v1.Health/Check", true},
		{"processlog denied", "/plugins.Integration/ProcessLog", false},
		{"unknown method denied", "/com.example.Unknown/Do", false},
		{"empty method denied", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := internalKeyPermitted(c.method); got != c.permitted {
				t.Fatalf("internalKeyPermitted(%q) = %v, want %v", c.method, got, c.permitted)
			}
		})
	}
}

// TestAuthFromContextInternalKeyDeny verifies that a bare internal-key on a
// non-allowlisted method is rejected with PermissionDenied. This is the core of
// the C2 fix: the internal-key branch must deny the data plane BEFORE the key
// is compared, so a valid internal key cannot open ProcessLog.
func TestAuthFromContextInternalKeyDeny(t *testing.T) {
	m := NewMiddlewares(testAuthService())

	deny := []struct {
		name   string
		method string
	}{
		{"processlog", "/plugins.Integration/ProcessLog"},
		{"unknown", "/com.example.Unknown/Do"},
		{"empty", ""},
	}

	for _, c := range deny {
		t.Run(c.name, func(t *testing.T) {
			err := m.authFromContext(internalKeyCtx(), c.method)
			if err == nil {
				t.Fatalf("authFromContext(internal-key, %q) = nil, want PermissionDenied", c.method)
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("authFromContext(internal-key, %q) error is not a gRPC status: %v", c.method, err)
			}
			if st.Code() != codes.PermissionDenied {
				t.Fatalf("authFromContext(internal-key, %q) code = %v, want PermissionDenied", c.method, st.Code())
			}
		})
	}
}

// TestAuthFromContextInternalKeyAllow verifies that a bare internal-key on the
// allowlisted health method passes (returns nil), proving the fix did not break
// the legitimate health path.
func TestAuthFromContextInternalKeyAllow(t *testing.T) {
	m := NewMiddlewares(testAuthService())

	err := m.authFromContext(internalKeyCtx(), "/grpc.health.v1.Health/Check")
	if err != nil {
		t.Fatalf("authFromContext(internal-key, health) = %v, want nil", err)
	}
}

// TestAuthFromContextDataPlaneUnchanged verifies the data-plane branches still
// authorize ProcessLog after the change: key/id/type (agents/collectors) and
// connection-key (federation). These must NOT be gated by the internal-key
// allowlist.
func TestAuthFromContextDataPlaneUnchanged(t *testing.T) {
	m := NewMiddlewares(testAuthService())
	const processLog = "/plugins.Integration/ProcessLog"

	// Collector key/id/type.
	err := m.authFromContext(keyIDTypeCtx("42", "collector-42-key", "collector"), processLog)
	if err != nil {
		t.Fatalf("authFromContext(collector key/id/type, ProcessLog) = %v, want nil", err)
	}

	// Agent key/id/type.
	err = m.authFromContext(keyIDTypeCtx("7", "agent-7-key", "agent"), processLog)
	if err != nil {
		t.Fatalf("authFromContext(agent key/id/type, ProcessLog) = %v, want nil", err)
	}

	// Federation connection-key.
	err = m.authFromContext(connectionKeyCtx("test-connection-key"), processLog)
	if err != nil {
		t.Fatalf("authFromContext(connection-key, ProcessLog) = %v, want nil", err)
	}

	// Negative controls: wrong collector key and wrong connection key must deny.
	if err := m.authFromContext(keyIDTypeCtx("42", "wrong-key", "collector"), processLog); err == nil {
		t.Fatalf("authFromContext(wrong collector key, ProcessLog) = nil, want PermissionDenied")
	}
	if err := m.authFromContext(connectionKeyCtx("wrong-connection-key"), processLog); err == nil {
		t.Fatalf("authFromContext(wrong connection-key, ProcessLog) = nil, want PermissionDenied")
	}
}
