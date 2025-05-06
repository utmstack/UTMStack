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
	Timestamp         string         `json:"@timestamp"`
	ID                string         `json:"id"`
	ParentID          *string        `json:"parentId,omitempty"`
	Status            int            `json:"status"`
	StatusLabel       string         `json:"statusLabel"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Name              string         `json:"name"`
	Category          string         `json:"category"`
	Severity          int            `json:"severity"`
	SeverityLabel     string         `json:"severityLabel"`
	Description       string         `json:"description"`
	Solution          string         `json:"solution"`
	Technique         string         `json:"technique"`
	Reference         []string       `json:"reference"`
	DataType          string         `json:"dataType"`
	Impact            *Impact        `json:"impact"`
	ImpactScore       int32          `json:"impactScore"`
	DataSource        string         `json:"dataSource"`
	Adversary         *Side          `json:"adversary"`
	Target            *Side          `json:"target"`
	Events            []*Event       `json:"events"`
	LastEvent         *Event         `json:"lastEvent"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"tagRulesApplied"`
	DeduplicatedBy    []string       `json:"deduplicatedBy"`
}

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type Impact struct {
	Confidentiality int32 `json:"confidentiality,omitempty"`
	Integrity       int32 `json:"integrity,omitempty"`
	Availability    int32 `json:"availability,omitempty"`
}

type Event struct {
	Id               string   `json:"id,omitempty"`
	Timestamp        string   `json:"timestamp,omitempty"`
	DeviceTime       string   `json:"deviceTime,omitempty"`
	DataType         string   `json:"dataType,omitempty"`
	DataSource       string   `json:"dataSource,omitempty"`
	TenantId         string   `json:"tenantId,omitempty"`
	TenantName       string   `json:"tenantName,omitempty"`
	Raw              string   `json:"raw,omitempty"`
	Log              []string `json:"log,omitempty"`
	Target           *Side    `json:"target,omitempty"`
	Origin           *Side    `json:"origin,omitempty"`
	Protocol         string   `json:"protocol,omitempty"`
	ConnectionStatus string   `json:"connectionStatus,omitempty"`
	StatusCode       int64    `json:"statusCode,omitempty"`
	ActionResult     string   `json:"actionResult,omitempty"`
	Action           string   `json:"action,omitempty"`
	Command          string   `json:"command,omitempty"`
	Severity         string   `json:"severity,omitempty"`
}

type Side struct {
	BytesSent        float64     `json:"bytesSent,omitempty"`
	BytesReceived    float64     `json:"bytesReceived,omitempty"`
	PackagesSent     int64       `json:"packagesSent,omitempty"`
	PackagesReceived int64       `json:"packagesReceived,omitempty"`
	Connections      int64       `json:"connections,omitempty"`
	UsedCpuPercent   int64       `json:"usedCpuPercent,omitempty"`
	UsedMemPercent   int64       `json:"usedMemPercent,omitempty"`
	TotalCpuUnits    int64       `json:"totalCpuUnits,omitempty"`
	TotalMem         int64       `json:"totalMem,omitempty"`
	Ip               string      `json:"ip,omitempty"`
	Host             string      `json:"host,omitempty"`
	User             string      `json:"user,omitempty"`
	Group            string      `json:"group,omitempty"`
	Port             int64       `json:"port,omitempty"`
	Domain           string      `json:"domain,omitempty"`
	Fqdn             string      `json:"fqdn,omitempty"`
	Mac              string      `json:"mac,omitempty"`
	Process          string      `json:"process,omitempty"`
	Geolocation      Geolocation `json:"geolocation,omitempty"`
	File             string      `json:"file,omitempty"`
	Path             string      `json:"path,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	Url              string      `json:"url,omitempty"`
	Email            string      `json:"email,omitempty"`
}

type Geolocation struct {
	GeolocationCountry     string  `json:"geolocationCountry,omitempty"`
	GeolocationCity        string  `json:"geolocationCity,omitempty"`
	GeolocationLatitude    float64 `json:"geolocationLatitude,omitempty"`
	GeolocationLongitude   float64 `json:"geolocationLongitude,omitempty"`
	GeolocationAsn         int64   `json:"geolocationAsn,omitempty"`
	GeolocationAso         string  `json:"geolocationAso,omitempty"`
	GeolocationCountryCode string  `json:"geolocationcCuntryCode,omitempty"`
	GeolocationAccuracy    int32   `json:"geolocationaAcuracy,omitempty"`
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
