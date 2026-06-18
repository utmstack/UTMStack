package dto

type RelatedLogsResponse struct {
	RuleMatched  bool     `json:"ruleMatched"`
	IndexPattern string   `json:"indexPattern"`
	IDs          []string `json:"ids"`
	Total        int      `json:"total"`
	Truncated    bool     `json:"truncated"`
	TimeFrom     string   `json:"timeFrom"`
	TimeTo       string   `json:"timeTo"`
}
