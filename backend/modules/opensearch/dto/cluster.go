package dto

import ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"

type ClusterStatusResponse struct {
	Health *ospkg.ClusterHealth `json:"health"`
}
