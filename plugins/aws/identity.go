package main

import (
	"strconv"

	"github.com/utmstack/UTMStack/plugins/shared/identity"
)

// CloudWatch carries no per-event ID, so the whole message must be hashed.
// groupKey must be ModuleGroup.Key(): log group and stream names are
// administrator-typed, so two tenants in separate AWS accounts can pick
// identical ones, and without it identical content collides across tenants.
func eventIdentity(groupKey, logGroup, streamName string, timestamp int64, message string) string {
	return identity.Hash(groupKey, logGroup, streamName, strconv.FormatInt(timestamp, 10), message)
}
