package mcp

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
)

// jsonschema-go's reflection-based inference has no special case for
// encoding.TextMarshaler, so a bare uuid.UUID field — structurally [16]byte —
// would otherwise advertise itself as an array of 16 integers instead of a
// UUID string. Every MCP client sends UUIDs as strings, so a regression here
// makes every ID-taking tool call fail schema validation before the handler
// ever runs. See uuidTypeSchemas in server.go.
func TestUUIDFieldsSchemaAsString(t *testing.T) {
	type withID struct {
		ID uuid.UUID `json:"id"`
	}

	schema, err := inputSchemaFor[withID]()
	if err != nil {
		t.Fatalf("inputSchemaFor: %v", err)
	}

	prop := schema.Properties["id"]
	if prop == nil {
		t.Fatalf("schema has no \"id\" property: %+v", schema)
	}
	if prop.Type != "string" {
		t.Fatalf("id property type = %q (Types=%v), want \"string\" — a uuid.UUID field must not fall back to the raw [16]byte array schema", prop.Type, prop.Types)
	}
}

// Pins the fix against the two real tool inputs that surfaced the bug.
func TestDashboardToolSchemasUseStringIDs(t *testing.T) {
	cases := []struct {
		name   string
		schema func() (*jsonschema.Schema, error)
	}{
		{"dashboardIDInput.id", func() (*jsonschema.Schema, error) { return inputSchemaFor[dashboardIDInput]() }},
		{"visualizationUpsertInput.dashboard_id", func() (*jsonschema.Schema, error) { return inputSchemaFor[visualizationUpsertInput]() }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			schema, err := c.schema()
			if err != nil {
				t.Fatalf("inputSchemaFor: %v", err)
			}
			for propName, prop := range schema.Properties {
				if propName != "id" && propName != "dashboard_id" {
					continue
				}
				if prop.Type != "string" {
					t.Fatalf("%s property type = %q, want \"string\"", propName, prop.Type)
				}
			}
		})
	}
}
