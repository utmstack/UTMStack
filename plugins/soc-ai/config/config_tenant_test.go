package config

import (
	"os"
	"path/filepath"
	"testing"
)

const tenantYAML = `
plugins:
    soc-ai:
        provider: openai
        model: gpt-4o
        url: https://api.openai.com/v1/chat/completions
        auth_type: bearer
        max_tokens: 4096
        auto_analyze: true
        tenants:
            tenant-a:
                provider: anthropic
                model: claude-opus-5
                url: https://api.anthropic.com/v1/messages
                auth_type: header
                auth_header_name: x-api-key
                max_tokens: 8192
                auto_analyze: false
`

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, pluginFile)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestReadConfigSplitsTheDefaultFromTheTenants(t *testing.T) {
	set := readConfig(writeConfig(t, tenantYAML), "", "http://backend:8080", "key")

	if set.Default == nil || set.Default.Provider != "openai" {
		t.Fatalf("default = %+v, want openai", set.Default)
	}
	own, ok := set.Tenants["tenant-a"]
	if !ok || own.Provider != "anthropic" {
		t.Fatalf("tenant-a = %+v, want anthropic", own)
	}

	// The platform fields reach every configuration, tenant or not: they say
	// where the backend is, which is not a tenant's business to choose.
	if own.Backend != "http://backend:8080" || own.InternalKey != "key" {
		t.Errorf("tenant lost the platform fields: %+v", own)
	}
	// Defaults are filled per entry, not inherited from the instance one.
	if own.MaxToolIterations != 12 {
		t.Errorf("MaxToolIterations = %d, want the default 12", own.MaxToolIterations)
	}
}

func TestGetConfigPrefersTheTenantAndFallsBack(t *testing.T) {
	set(readConfig(writeConfig(t, tenantYAML), "", "http://backend:8080", "key"))

	if got := GetConfig("tenant-a"); got.Provider != "anthropic" {
		t.Errorf("tenant-a = %q, want anthropic", got.Provider)
	}
	if got := GetConfig("tenant-without-one"); got.Provider != "openai" {
		t.Errorf("unconfigured tenant = %q, want the instance default", got.Provider)
	}
	if got := GetConfig(""); got.Provider != "openai" {
		t.Errorf("empty tenant = %q, want the instance default", got.Provider)
	}
}

// The override replaces the default outright: auto_analyze is true on the
// instance and false for tenant-a, and it must not leak across.
func TestTenantOverrideDoesNotInheritFieldByField(t *testing.T) {
	set(readConfig(writeConfig(t, tenantYAML), "", "b", "k"))

	if !GetConfig("").AutoAnalyze {
		t.Error("the instance default lost auto_analyze")
	}
	if GetConfig("tenant-a").AutoAnalyze {
		t.Error("tenant-a inherited auto_analyze from the default")
	}
}

func TestReadConfigWithNoTenantsBlock(t *testing.T) {
	set := readConfig(writeConfig(t, `
plugins:
    soc-ai:
        provider: openai
        model: gpt-4o
        url: https://x
`), "", "b", "k")

	if !set.Default.ModuleActive {
		t.Fatal("default not active")
	}
	if len(set.Tenants) != 0 {
		t.Fatalf("tenants = %v, want none", set.Tenants)
	}
	if got := GetConfig("anyone"); got == nil {
		t.Fatal("a tenant got nothing when there is a usable default")
	}
}
