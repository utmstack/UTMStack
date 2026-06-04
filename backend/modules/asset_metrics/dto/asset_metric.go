package dto

type AssetMetricRequest struct {
	ID        string `json:"id"        binding:"omitempty"`
	AssetName string `json:"assetName" binding:"required"`
	Metric    string `json:"metric"    binding:"required"`
	Amount    int64  `json:"amount"    binding:"required"`
}

type AssetMetricResponse struct {
	ID        string `json:"id"`
	AssetName string `json:"assetName"`
	Metric    string `json:"metric"`
	Amount    int64  `json:"amount"`
}
