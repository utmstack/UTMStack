package main

import "github.com/utmstack/UTMStack/plugins/shared/identity"

// CrowdStrike caps the appId query parameter at 32 alphanumeric characters.
// Truncating to 32 hex characters keeps 128 bits, ample to keep units distinct.
const appIDLength = 32

// deriveAppID derives the CrowdStrike appId for one owned unit. An appId names a
// server-side subscription carrying its own offset, so it must be stable per
// unit: a distinct appId per process opens an independent copy of the whole feed
// rather than distributing it. groupKey must be passed through unmodified.
func deriveAppID(groupKey string) string {
	return identity.Hash(groupKey)[:appIDLength]
}
