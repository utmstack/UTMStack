package upgrade

import "github.com/google/uuid"

// Default tenant. v12 seeds this row at bootstrap; every converted v11 record
// belongs to it.
const defaultTenantIDStr = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var defaultTenantID = uuid.MustParse(defaultTenantIDStr)

// Display names for the default tenant in the two places it appears.
const (
	// DB name (tenant table).
	defaultTenantDBName = "UTMStack"
	// Event-engine tenants.yaml name; must match the engine's DefaultTenantName.
	eventEngineTenantName = "Default"
)

// v11 authority name -> v12 built-in role UUIDs (seeded by v12 bootstrap).
const (
	roleAdminID   = "00000000-0000-0000-0000-00000000a001"
	roleAnalystID = "00000000-0000-0000-0000-00000000a002"
	roleViewerID  = "00000000-0000-0000-0000-00000000a003"
)

// UUID v5 namespaces, one per entity. Converting a v11 numeric ID to a v12
// UUID must be deterministic (idempotent re-runs) and must not collide across
// tables (a user and an incident can share the same numeric ID). Values come
// from the original MigrationTool; do not change them.
var (
	userNamespace            = uuid.MustParse("8f3c9a1e-7b2d-4e6a-9c1f-5d8e0a3b7c21")
	configNamespace          = uuid.MustParse("04f5f62e-5ec0-4042-825b-2c7b8a38ec73")
	alertTagNamespace        = uuid.MustParse("402218d7-f216-56d0-a6bc-be7414c8aff9")
	alertTagRuleNamespace    = uuid.MustParse("d65b0f32-73f4-5340-acf5-ae4df2dbaba0")
	apiKeyNamespace          = uuid.MustParse("1f52d62b-6fdb-b2db-ca5e-fd519a49fafd")
	incidentNamespace        = uuid.MustParse("ffa82061-9bd3-56c0-8f37-e34e33e2f9d0")
	incidentAlertNamespace   = uuid.MustParse("b685a259-ebac-589d-9425-cfc09cf2e9b9")
	incidentHistoryNamespace = uuid.MustParse("c7c9699c-a814-5e85-ba9d-ad11ddac670d")
	incidentNoteNamespace    = uuid.MustParse("351bf618-4905-5df6-bf66-2e7280a205d0")
)
