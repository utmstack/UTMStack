package schema

import (
	"fmt"
	"strings"
)

// Valid LLM classification values
const (
	ClassificationPossibleIncident     = "possible incident"
	ClassificationPossibleFalsePositive = "possible false positive"
	ClassificationStandardAlert        = "standard alert"
)

// ValidClassifications maps valid classification values for quick lookup
var ValidClassifications = map[string]string{
	"possible incident":      ClassificationPossibleIncident,
	"possible false positive": ClassificationPossibleFalsePositive,
	"standard alert":         ClassificationStandardAlert,
}

// SearchDetailsRequest is used for searching alert details via the backend API
type SearchDetailsRequest []struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// GPTRequest represents a request to the LLM API (OpenAI compatible)
type GPTRequest struct {
	Model     string       `json:"model"`
	Messages  []GPTMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"` // Required for some providers (e.g., Anthropic)
}

// GPTMessage represents a message in the LLM conversation
type GPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GPTResponse represents the response from the LLM API
type GPTResponse struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int         `json:"created"`
	Model             string      `json:"model"`
	Choices           []GPTChoice `json:"choices"`
	Usage             GPTUsage    `json:"usage"`
	SystemFingerprint string      `json:"system_fingerprint"`
}

// GPTChoice represents a choice in the LLM response
type GPTChoice struct {
	Index        int        `json:"index"`
	Message      GPTMessage `json:"message"`
	LogProbs     string     `json:"logprobs"`
	FinishReason string     `json:"finish_reason"`
}

// GPTUsage represents token usage information
type GPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model     string              `json:"model"`
	System    string              `json:"system,omitempty"`
	Messages  []AnthropicMessage  `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
}

// AnthropicMessage represents a message in the Anthropic conversation
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response from Anthropic API
type AnthropicResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []AnthropicContent `json:"content"`
	Model   string             `json:"model"`
}

// AnthropicContent represents content in Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// GPTAlertResponse represents the structured response from LLM analysis
type GPTAlertResponse struct {
	Timestamp      string     `json:"@timestamp,omitempty"`
	Status         string     `json:"status,omitempty"`
	Severity       int        `json:"severity,omitempty"`
	Category       string     `json:"category,omitempty"`
	AlertName      string     `json:"alertName,omitempty"`
	ActivityID     string     `json:"activityId,omitempty"`
	Classification string     `json:"classification,omitempty"`
	Reasoning      []string   `json:"reasoning,omitempty"`
	NextSteps      []NextStep `json:"nextSteps,omitempty"`
}

// NextStep represents a recommended action step
type NextStep struct {
	Step    int    `json:"step"`
	Action  string `json:"action"`
	Details string `json:"details"`
}

// Validate checks that the LLM response has valid required fields
// and normalizes the classification to lowercase
func (r *GPTAlertResponse) Validate() error {
	// Normalize and validate classification
	classification := strings.ToLower(strings.TrimSpace(r.Classification))
	if classification == "" {
		return fmt.Errorf("classification is required")
	}

	normalized, valid := ValidClassifications[classification]
	if !valid {
		return fmt.Errorf("invalid classification '%s': must be one of: %s, %s, %s",
			r.Classification,
			ClassificationPossibleIncident,
			ClassificationPossibleFalsePositive,
			ClassificationStandardAlert)
	}
	r.Classification = normalized

	// Validate reasoning exists
	if len(r.Reasoning) == 0 {
		return fmt.Errorf("reasoning is required: LLM must provide at least one reason")
	}

	// Filter out empty reasoning strings
	validReasons := make([]string, 0, len(r.Reasoning))
	for _, reason := range r.Reasoning {
		if trimmed := strings.TrimSpace(reason); trimmed != "" {
			validReasons = append(validReasons, trimmed)
		}
	}
	if len(validReasons) == 0 {
		return fmt.Errorf("reasoning is required: all provided reasons were empty")
	}
	r.Reasoning = validReasons

	return nil
}

// ChangeAlertStatus is the request body for changing alert status
type ChangeAlertStatus struct {
	AlertIDs          []string `json:"alertIds"`
	DataSource        string   `json:"dataSource"`
	StatusObservation string   `json:"statusObservation"`
	Status            int      `json:"status"`
}

// CreateNewIncidentRequest is the request body for creating a new incident
type CreateNewIncidentRequest struct {
	IncidentName        string    `json:"incidentName"`
	IncidentDescription string    `json:"incidentDescription"`
	IncidentAssignedTo  string    `json:"incidentAssignedTo"`
	AlertList           AlertList `json:"alertList"`
}

// AlertList represents a list of alerts for incident operations
type AlertList []struct {
	AlertID       string `json:"alertId"`
	AlertName     string `json:"alertName"`
	AlertStatus   int    `json:"alertStatus"`
	AlertSeverity int    `json:"alertSeverity"`
}

// IncidentResp represents an incident response from the API
type IncidentResp struct {
	ID                  int    `json:"id"`
	IncidentName        string `json:"incidentName"`
	IncidentDescription string `json:"incidentDescription"`
	IncidentStatus      string `json:"incidentStatus"`
	IncidentAssignedTo  string `json:"incidentAssignedTo"`
	IncidentSeverity    int    `json:"incidentSeverity"`
	IncidentCreatedDate string `json:"incidentCreatedDate"`
	IncidentSolution    string `json:"incidentSolution"`
}

// AddNewAlertToIncidentRequest is the request body for adding alerts to an incident
type AddNewAlertToIncidentRequest struct {
	IncidentId int       `json:"incidentId"`
	AlertList  AlertList `json:"alertList"`
}

// UpdateDocRequest represents an Elasticsearch update by query request
type UpdateDocRequest struct {
	Query  Query  `json:"query"`
	Script Script `json:"script"`
}

// Query represents an Elasticsearch query
type Query struct {
	Bool `json:"bool"`
}

// Bool represents an Elasticsearch bool query
type Bool struct {
	Must []Must `json:"must"`
}

// Must represents a must clause in Elasticsearch bool query
type Must struct {
	Match Match `json:"match"`
}

// Match represents a match clause in Elasticsearch query
type Match struct {
	ActivityID string `json:"activityId"`
}

// Script represents an Elasticsearch script for updates
type Script struct {
	Source string           `json:"source"`
	Lang   string           `json:"lang"`
	Params GPTAlertResponse `json:"params"`
}
