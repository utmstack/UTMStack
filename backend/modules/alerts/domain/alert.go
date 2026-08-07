package domain

import "encoding/json"

type IncidentDetail struct {
	IncidentName string `json:"incidentName,omitempty"`
	IncidentID   any    `json:"incidentId,omitempty"`
	CreationDate string `json:"creationDate,omitempty"`
	CreatedBy    string `json:"createdBy,omitempty"`
	Source       string `json:"source,omitempty"`
}

type Geolocation struct {
	Country     string  `json:"country,omitempty"`
	City        string  `json:"city,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	ASN         int64   `json:"asn,omitempty"`
	ASO         string  `json:"aso,omitempty"`
	CountryCode string  `json:"countryCode,omitempty"`
	Accuracy    int     `json:"accuracy,omitempty"`
}

type DiskInfo struct {
	Name        string `json:"name,omitempty"`
	TotalSpace  int64  `json:"totalSpace,omitempty"`
	UsedPercent int    `json:"usedPercent,omitempty"`
}

type Side struct {
	IP                        string       `json:"ip,omitempty"`
	Host                      string       `json:"host,omitempty"`
	User                      string       `json:"user,omitempty"`
	Group                     string       `json:"group,omitempty"`
	Port                      int          `json:"port,omitempty"`
	Domain                    string       `json:"domain,omitempty"`
	Mac                       string       `json:"mac,omitempty"`
	URL                       string       `json:"url,omitempty"`
	CIDR                      string       `json:"cidr,omitempty"`
	Name                      string       `json:"name,omitempty"`
	Geolocation               *Geolocation `json:"geolocation,omitempty"`
	BytesSent                 float64      `json:"bytesSent,omitempty"`
	BytesReceived             float64      `json:"bytesReceived,omitempty"`
	PackagesSent              int64        `json:"packagesSent,omitempty"`
	PackagesReceived          int64        `json:"packagesReceived,omitempty"`
	CertificateFingerprint    string       `json:"certificateFingerprint,omitempty"`
	JA3Fingerprint            string       `json:"ja3Fingerprint,omitempty"`
	JARMFingerprint           string       `json:"jarmFingerprint,omitempty"`
	SSHBanner                 string       `json:"sshBanner,omitempty"`
	SSHFingerprint            string       `json:"sshFingerprint,omitempty"`
	Cookie                    string       `json:"cookie,omitempty"`
	JabberID                  string       `json:"jabberId,omitempty"`
	Email                     string       `json:"email,omitempty"`
	DKIM                      string       `json:"dkim,omitempty"`
	DKIMSignature             string       `json:"dkimSignature,omitempty"`
	EmailAddress              string       `json:"emailAddress,omitempty"`
	EmailBody                 string       `json:"emailBody,omitempty"`
	EmailDisplayName          string       `json:"emailDisplayName,omitempty"`
	EmailSubject              string       `json:"emailSubject,omitempty"`
	EmailThreadIndex          string       `json:"emailThreadIndex,omitempty"`
	EmailXMailer              string       `json:"emailXMailer,omitempty"`
	WhoisRegistrant           string       `json:"whoisRegistrant,omitempty"`
	WhoisRegistrar            string       `json:"whoisRegistrar,omitempty"`
	Process                   string       `json:"process,omitempty"`
	ProcessState              string       `json:"processState,omitempty"`
	Command                   string       `json:"command,omitempty"`
	WindowsScheduledTask      string       `json:"windowsScheduledTask,omitempty"`
	WindowsServiceDisplayName string       `json:"windowsServiceDisplayName,omitempty"`
	WindowsServiceName        string       `json:"windowsServiceName,omitempty"`
	File                      string       `json:"file,omitempty"`
	Path                      string       `json:"path,omitempty"`
	Filename                  string       `json:"filename,omitempty"`
	SizeInBytes               string       `json:"sizeInBytes,omitempty"`
	MimeType                  string       `json:"mimeType,omitempty"`
	Hash                      string       `json:"hash,omitempty"`
	Authentihash              string       `json:"authentihash,omitempty"`
	CDHash                    string       `json:"cdhash,omitempty"`
	MD5                       string       `json:"md5,omitempty"`
	SHA1                      string       `json:"sha1,omitempty"`
	SHA224                    string       `json:"sha224,omitempty"`
	SHA256                    string       `json:"sha256,omitempty"`
	SHA384                    string       `json:"sha384,omitempty"`
	SHA3224                   string       `json:"sha3224,omitempty"`
	SHA3256                   string       `json:"sha3256,omitempty"`
	SHA3384                   string       `json:"sha3384,omitempty"`
	SHA3512                   string       `json:"sha3512,omitempty"`
	SHA512                    string       `json:"sha512,omitempty"`
	SHA512224                 string       `json:"sha512224,omitempty"`
	SHA512256                 string       `json:"sha512256,omitempty"`
	Hex                       string       `json:"hex,omitempty"`
	Base64                    string       `json:"base64,omitempty"`
	OperatingSystem           string       `json:"operatingSystem,omitempty"`
	ChromeExtension           string       `json:"chromeExtension,omitempty"`
	MobileAppID               string       `json:"mobileAppId,omitempty"`
	CPE                       string       `json:"cpe,omitempty"`
	CVE                       string       `json:"cve,omitempty"`
	Malware                   string       `json:"malware,omitempty"`
	MalwareFamily             string       `json:"malwareFamily,omitempty"`
	MalwareType               string       `json:"malwareType,omitempty"`
	PGPPrivateKey             string       `json:"pgpPrivateKey,omitempty"`
	PGPPublicKey              string       `json:"pgpPublicKey,omitempty"`
	Connections               int64        `json:"connections,omitempty"`
	UsedCPUPercent            int          `json:"usedCpuPercent,omitempty"`
	UsedMemPercent            int          `json:"usedMemPercent,omitempty"`
	TotalCPUUnits             int          `json:"totalCpuUnits,omitempty"`
	TotalMem                  int64        `json:"totalMem,omitempty"`
	Disks                     []DiskInfo   `json:"disks,omitempty"`
}

type Impact struct {
	Confidentiality int `json:"confidentiality,omitempty"`
	Integrity       int `json:"integrity,omitempty"`
	Availability    int `json:"availability,omitempty"`
}

type AlertEvent struct {
	ID        string `json:"id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type UtmAlert struct {
	TenantID          string          `json:"tenantId,omitempty"`
	TenantName        string          `json:"tenantName,omitempty"`
	Timestamp         string          `json:"@timestamp,omitempty"`
	LastUpdate        string          `json:"lastUpdate,omitempty"`
	ID                string          `json:"id,omitempty"`
	ParentID          string          `json:"parentId,omitempty"`
	Name              string          `json:"name,omitempty"`
	Category          string          `json:"category,omitempty"`
	Technique         string          `json:"technique,omitempty"`
	Description       string          `json:"description,omitempty"`
	Solution          string          `json:"solution,omitempty"`
	DataType          string          `json:"dataType,omitempty"`
	DataSource        string          `json:"dataSource,omitempty"`
	Severity          AlertSeverity   `json:"severity,omitempty"`
	Status            AlertStatus     `json:"status,omitempty"`
	StatusObservation string          `json:"statusObservation,omitempty"`
	Impact            *Impact         `json:"impact,omitempty"`
	ImpactScore       int             `json:"impactScore,omitempty"`
	Adversary         *Side           `json:"adversary,omitempty"`
	Target            *Side           `json:"target,omitempty"`
	LastEvent         json.RawMessage `json:"lastEvent,omitempty"`
	Events            []AlertEvent    `json:"events,omitempty"`
	References        []string        `json:"references,omitempty"`
	DeduplicateBy     []string        `json:"deduplicateBy,omitempty"`
	GroupBy           []string        `json:"groupBy,omitempty"`
	Errors            []string        `json:"errors,omitempty"`
	IsIncident        bool            `json:"isIncident,omitempty"`
	IncidentDetail    *IncidentDetail `json:"incidentDetail,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	TagRulesApplied   []string        `json:"tagRulesApplied,omitempty"`
	Notes             string          `json:"notes,omitempty"`
	Assignee          string          `json:"assignee,omitempty"`
	History           []AlertHistory  `json:"history,omitempty"`
}

type AlertHistory struct {
	User      string `json:"user,omitempty"`
	Action    string `json:"action,omitempty"`
	NewValue  string `json:"newValue,omitempty"`  // compact JSON of the changed fields
	Timestamp string `json:"timestamp,omitempty"` // RFC3339
}

type AlertAction string

const (
	ActionUpdateStatus   AlertAction = "UPDATE_STATUS"
	ActionUpdateTags     AlertAction = "UPDATE_TAGS"
	ActionUpdateNotes    AlertAction = "UPDATE_NOTES"
	ActionUpdateSolution AlertAction = "UPDATE_SOLUTION"
	ActionMarkAsIncident AlertAction = "MARK_AS_INCIDENT"
	ActionUpdateAssignee AlertAction = "UPDATE_ASSIGNEE"
)

type AlertSeverity string

const (
	SeverityLow    AlertSeverity = "low"
	SeverityMedium AlertSeverity = "medium"
	SeverityHigh   AlertSeverity = "high"
)

func IsValidSeverity(s AlertSeverity) bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh:
		return true
	default:
		return false
	}
}

type AlertStatus string

const (
	AlertStatusAutomaticReview AlertStatus = "Automatic review"
	AlertStatusOpen            AlertStatus = "Open"
	AlertStatusInReview        AlertStatus = "In review"
	AlertStatusCompleted       AlertStatus = "Completed"
	AlertStatusMerged          AlertStatus = "Merged"
)

// StatusFromCode maps the numeric status the HTTP API and the incidents module
// still speak onto the stored value. It is a boundary shim and nothing below it
// knows the codes: they are a contract with callers, not a model.
func StatusFromCode(code int) AlertStatus {
	switch code {
	case 0:
		return AlertStatusMerged
	case 1:
		return AlertStatusAutomaticReview
	case 2:
		return AlertStatusOpen
	case 3:
		return AlertStatusInReview
	case 5:
		return AlertStatusCompleted
	default:
		return ""
	}
}

func IsValid(s AlertStatus) bool {
	switch s {
	case AlertStatusAutomaticReview, AlertStatusOpen, AlertStatusInReview,
		AlertStatusCompleted, AlertStatusMerged:
		return true
	default:
		return false
	}
}
