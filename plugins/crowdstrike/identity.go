package main

import (
	"strconv"

	"github.com/utmstack/UTMStack/plugins/shared/identity"
)

// eventIdentity derives a deterministic identity for one Falcon streaming
// event. event.Metadata.Offset is a per-subscription sequence number, unique
// only within a tenant, so tenant and group must be hashed too: otherwise two
// tenants each running a group named "prod" would collide.
func eventIdentity(tenantId, groupKey, streamID string, offset uint64) string {
	return identity.Hash(tenantId, groupKey, streamID, strconv.FormatUint(offset, 10))
}
