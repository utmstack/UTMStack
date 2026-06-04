// Package baseline carries the shared ModuleKind helpers (the Defaults
// no-op struct and the Azure-EventHub config-key factory) that per-kind
// packages embed and reuse. It lives one level below modulekinds so the
// parent modulekinds package can import every kind for RegisterAll without
// dragging the per-kind packages into an import cycle.
package baseline

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
)

// ValidateHTTPClient is the shared client used by per-kind validators that
// reach out to external provider APIs. 10s matches the plugin's legacy timeout.
var ValidateHTTPClient = &http.Client{Timeout: 10 * time.Second}

// ConfigValue looks up a config row by its ConfKey and returns the value, or
// the empty string when the key is absent. Per-kind validators use this to pull
// the handful of keys they care about without writing a switch each time.
func ConfigValue(configs []domain.UtmModuleGroupConfiguration, key string) string {
	for _, c := range configs {
		if c.ConfKey == key {
			return c.ConfValue
		}
	}
	return ""
}

// RequireFields returns the first "X is required" error encountered when
// scanning the supplied configs for the listed keys. The pairs are
// (confKey, humanLabel) — humanLabel is what appears in the error message and
// should match the label the panel shows the user.
//
// Pass an even number of arguments: "aws_default_region", "Default Region",
// "aws_access_key_id", "Access Key", … The function panics on an odd count so
// callers see the bug at compile-cycle test time, not at runtime.
func RequireFields(configs []domain.UtmModuleGroupConfiguration, moduleLabel string, pairs ...string) error {
	if len(pairs)%2 != 0 {
		panic("baseline.RequireFields: pairs must have an even number of elements (key,label,…)")
	}
	for i := 0; i < len(pairs); i += 2 {
		key, label := pairs[i], pairs[i+1]
		if ConfigValue(configs, key) == "" {
			return fmt.Errorf("%s is required in %s configuration", label, moduleLabel)
		}
	}
	return nil
}

// Defaults provides no-op implementations of the optional ModuleKind methods so
// per-kind packages can embed it and only override what they actually need
// (typically Name and ConfigurationKeys). It carries the kind name so the
// embedded Name() returns the right value without each kind reimplementing it.
type Defaults struct {
	KindName string
}

// Name returns the embedded kind name. Per-kind structs are expected to embed
// Defaults{KindName: "..."} so this method satisfies ModuleKind.Name without
// extra boilerplate.
func (d Defaults) Name() string { return d.KindName }

// ConfigurationKeys returns an empty list. Kinds that have default config
// rows override this method.
func (Defaults) ConfigurationKeys(_ int64) []domain.ModuleConfigurationKey {
	return nil
}

// CheckRequirements returns no checks. Kinds with preflight requirements
// override this.
func (Defaults) CheckRequirements(_ context.Context, _ int64) ([]domain.ModuleRequirement, error) {
	return nil, nil
}

// ValidateConfiguration accepts any config. Kinds with custom validation
// override this; the legacy default was likewise "return true".
func (Defaults) ValidateConfiguration(_ context.Context, _ *domain.UtmModule, _ []domain.UtmModuleGroupConfiguration) error {
	return nil
}

// UpdateModule is a post-update hook. The framework already publishes the
// module state to event-processor on every config write; this method lets a
// kind run kind-specific side effects on top of that (cache invalidation,
// connector restarts, …). The default is a no-op.
func (Defaults) UpdateModule(_ context.Context, _ string, _ []domain.UtmModuleGroupConfiguration) error {
	return nil
}

// EventHubKeys is the four-row config shape Azure-streamed log sources share
// (Azure itself, plus several appliances that forward through an Azure Event
// Hub: CISCO_SWITCH, DECEPTIVE_BYTES, FIRE_POWER, GITHUB, MACOS, MIKROTIK,
// PALO_ALTO, SONIC_WALL, UFW). All eleven legacy modules used the same keys
// but varied the user-facing labels; we keep that quirk so the panel UI shows
// the labels each integration's docs expect.
func EventHubKeys(groupID int64, ehName, cgName, scName, conName string) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "eventHubConnection",
			ConfName: ehName, ConfDescription: "Configure the event hub connection",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "consumerGroup",
			ConfName: cgName, ConfDescription: "Configure the consumer group",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "storageContainer",
			ConfName: scName, ConfDescription: "Configure the storage container",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "storageConnection",
			ConfName: conName, ConfDescription: "Configure the storage connection",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
	}
}
