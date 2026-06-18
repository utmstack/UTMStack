package dto

// IndexInfo is one entry of the index listing (_cat/indices shape).
type IndexInfo struct {
	Name      string `json:"index"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	DocsCount string `json:"docs.count"`
	StoreSize string `json:"store.size"`
}

// ClusterHealth holds the key fields from the _cluster/health response.
type ClusterHealth struct {
	ClusterName       string `json:"cluster_name"`
	Status            string `json:"status"`
	NumberOfNodes     int    `json:"number_of_nodes"`
	NumberOfDataNodes int    `json:"number_of_data_nodes"`
	ActiveShards      int    `json:"active_shards"`
}

// IndexPropertyType is a flattened index mapping field name + type.
type IndexPropertyType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
