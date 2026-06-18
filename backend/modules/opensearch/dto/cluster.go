package dto

type ClusterStatusResponse struct {
	Health *ClusterHealth `json:"health"`
}
