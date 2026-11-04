package main

import (
	"encoding/json"

	"github.com/utmstack/UTMStack/plugins/shared/identity"
)

// graphIdentity makes a re-read (window overlap) recognisable downstream:
// the same event always hashes to the same id. lastUpdate covers
// in-place-updated resources (alerts, incidents): each mutation is new.
func graphIdentity(utmTenantId, groupKey, source, objectID, lastUpdate string) string {
	return identity.Hash(utmTenantId, groupKey, source, objectID, lastUpdate)
}

// huntingIdentity identifies a row by its stable key; rows without one fall
// back to a hash of their content so distinct rows never collide.
func huntingIdentity(utmTenantId, groupKey, table string, row map[string]any) string {
	key, _ := row["RowKey"].(string)
	if key == "" {
		id, _ := row["id"].(string)
		key = id
	}
	if key == "" {
		b, _ := json.Marshal(row)
		key = string(b)
	}
	return identity.Hash(utmTenantId, groupKey, "hunting", table, key)
}
