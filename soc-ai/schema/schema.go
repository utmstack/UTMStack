package schema

type SearchResult struct{}

type SeachRequest struct{}

type SearchDetailsRequest []struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type GPTRequest struct {
	Model    string       `json:"model"`
	Messages []GPTMessage `json:"messages"`
}

type GPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GPTResponse struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int         `json:"created"`
	Model             string      `json:"model"`
	Choices           []GPTChoice `json:"choices"`
	Usage             GPTUsage    `json:"usage"`
	SystemFingerprint string      `json:"system_fingerprint"`
}

type GPTChoice struct {
	Index        int        `json:"index"`
	Message      GPTMessage `json:"message"`
	LogProbs     string     `json:"logprobs"`
	FinishReason string     `json:"finish_reason"`
}

type GPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

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

type NextStep struct {
	Step    int    `json:"step"`
	Action  string `json:"action"`
	Details string `json:"details"`
}

type ChangeAlertStatus struct {
	AlertIDs          []string `json:"alertIds"`
	StatusObservation string   `json:"statusObservation"`
	Status            int      `json:"status"`
}

type CreateNewIncidentRequest struct {
	IncidentName        string    `json:"incidentName"`
	IncidentDescription string    `json:"incidentDescription"`
	IncidentAssignedTo  string    `json:"incidentAssignedTo"`
	AlertList           AlertList `json:"alertList"`
}

type AlertList []struct {
	AlertID       string `json:"alertId"`
	AlertName     string `json:"alertName"`
	AlertStatus   int    `json:"alertStatus"`
	AlertSeverity int    `json:"alertSeverity"`
}

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

type AddNewAlertToIncidentRequest struct {
	IncidenId int       `json:"incidentId"`
	AlertList AlertList `json:"alertList"`
}

type AlertDetails []Alert

type Alert struct {
	ID                string         `json:"id"`
	Severity          int            `json:"severity"`
	TagRulesApplied   []int          `json:"TagRulesApplied,omitempty"`
	SeverityLabel     string         `json:"severityLabel"`
	Notes             string         `json:"notes"`
	DataType          string         `json:"dataType"`
	Description       string         `json:"description"`
	StatusLabel       string         `json:"statusLabel"`
	Tactic            string         `json:"tactic"`
	Tags              []string       `json:"tags"`
	Reference         []string       `json:"reference"`
	Protocol          string         `json:"protocol"`
	Timestamp         string         `json:"@timestamp"`
	Solution          string         `json:"solution"`
	StatusObservation string         `json:"statusObservation"`
	Name              string         `json:"name"`
	IsIncident        bool           `json:"isIncident"`
	Category          string         `json:"category"`
	DataSource        string         `json:"dataSource"`
	Logs              []string       `json:"logs"`
	Status            int            `json:"status"`
	Destination       Destination    `json:"destination"`
	Source            Source         `json:"source"`
	IncidentDetails   IncidentDetail `json:"incidentDetail"`
}

type Destination struct {
	Country             string    `json:"country"`
	AccuracyRadius      int       `json:"accuracyRadius"`
	City                string    `json:"city"`
	IP                  string    `json:"ip,omitempty"`
	Port                int       `json:"port"`
	CountryCode         string    `json:"countryCode"`
	IsAnonymousProxy    bool      `json:"isAnonymousProxy"`
	Host                string    `json:"host,omitempty"`
	Coordinates         []float64 `json:"coordinates"`
	IsSatelliteProvider bool      `json:"isSatelliteProvider"`
	Aso                 string    `json:"aso"`
	Asn                 int       `json:"asn"`
	User                string    `json:"user"`
}

type Source struct {
	Country             string    `json:"country"`
	AccuracyRadius      int       `json:"accuracyRadius"`
	City                string    `json:"city"`
	IP                  string    `json:"ip,omitempty"`
	Coordinates         []float64 `json:"coordinates"`
	Port                int       `json:"port"`
	CountryCode         string    `json:"countryCode"`
	IsAnonymousProxy    bool      `json:"isAnonymousProxy"`
	IsSatelliteProvider bool      `json:"isSatelliteProvider"`
	Host                string    `json:"host,omitempty"`
	Aso                 string    `json:"aso"`
	Asn                 int       `json:"asn"`
	User                string    `json:"user"`
}

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	Source       string `json:"source"`
	CreationDate string `json:"creationDate"`
	IncidentName string `json:"incidentName,omitempty"`
	IncidentID   string `json:"incidentId,omitempty"`
}

type UpdateDocRequest struct {
	Query  Query  `json:"query"`
	Script Script `json:"script"`
}

type Query struct {
	Bool `json:"bool"`
}

type Bool struct {
	Must []Must `json:"must"`
}

type Must struct {
	Match Match `json:"match"`
}

type Match struct {
	ActivityID string `json:"activityId"`
}

type Script struct {
	Source string           `json:"source"`
	Lang   string           `json:"lang"`
	Params GPTAlertResponse `json:"params"`
}
