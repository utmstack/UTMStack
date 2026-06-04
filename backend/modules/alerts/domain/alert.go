package domain

type IncidentDetail struct {
	IncidentName string `json:"incidentName,omitempty"`
	IncidentID   any    `json:"incidentId,omitempty"`
	CreationDate string `json:"creationDate,omitempty"`
	CreatedBy    string `json:"createdBy,omitempty"`
	Source       string `json:"source,omitempty"`
}

// Geolocation holds geographic information for a Side.
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

// DiskInfo holds disk usage information for a Side.
type DiskInfo struct {
	Name        string `json:"name,omitempty"`
	TotalSpace  int64  `json:"totalSpace,omitempty"`
	UsedPercent int    `json:"usedPercent,omitempty"`
}

// Side represents one endpoint (adversary or target) in an alert.
// Fields map 1:1 to Java's Side class for full JSON round-trip parity.
type Side struct {
	// Network identification
	IP     string `json:"ip,omitempty"`
	Host   string `json:"host,omitempty"`
	User   string `json:"user,omitempty"`
	Group  string `json:"group,omitempty"`
	Port   int    `json:"port,omitempty"`
	Domain string `json:"domain,omitempty"`
	Mac    string `json:"mac,omitempty"`
	URL    string `json:"url,omitempty"`
	CIDR   string `json:"cidr,omitempty"`

	// Legacy field (kept for backward compat — mapped from "name" in older docs)
	Name string `json:"name,omitempty"`

	// Geolocation
	Geolocation *Geolocation `json:"geolocation,omitempty"`

	// Network traffic
	BytesSent        float64 `json:"bytesSent,omitempty"`
	BytesReceived    float64 `json:"bytesReceived,omitempty"`
	PackagesSent     int64   `json:"packagesSent,omitempty"`
	PackagesReceived int64   `json:"packagesReceived,omitempty"`

	// Certificate and fingerprint
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	JA3Fingerprint         string `json:"ja3Fingerprint,omitempty"`
	JARMFingerprint        string `json:"jarmFingerprint,omitempty"`
	SSHBanner              string `json:"sshBanner,omitempty"`
	SSHFingerprint         string `json:"sshFingerprint,omitempty"`

	// Web
	Cookie    string `json:"cookie,omitempty"`
	JabberID  string `json:"jabberId,omitempty"`

	// Email
	Email              string `json:"email,omitempty"`
	DKIM               string `json:"dkim,omitempty"`
	DKIMSignature      string `json:"dkimSignature,omitempty"`
	EmailAddress       string `json:"emailAddress,omitempty"`
	EmailBody          string `json:"emailBody,omitempty"`
	EmailDisplayName   string `json:"emailDisplayName,omitempty"`
	EmailSubject       string `json:"emailSubject,omitempty"`
	EmailThreadIndex   string `json:"emailThreadIndex,omitempty"`
	EmailXMailer       string `json:"emailXMailer,omitempty"`

	// WHOIS
	WhoisRegistrant string `json:"whoisRegistrant,omitempty"`
	WhoisRegistrar  string `json:"whoisRegistrar,omitempty"`

	// Process
	Process                  string `json:"process,omitempty"`
	ProcessState             string `json:"processState,omitempty"`
	Command                  string `json:"command,omitempty"`
	WindowsScheduledTask     string `json:"windowsScheduledTask,omitempty"`
	WindowsServiceDisplayName string `json:"windowsServiceDisplayName,omitempty"`
	WindowsServiceName       string `json:"windowsServiceName,omitempty"`

	// File
	File        string `json:"file,omitempty"`
	Path        string `json:"path,omitempty"`
	Filename    string `json:"filename,omitempty"`
	SizeInBytes string `json:"sizeInBytes,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`

	// Hashes
	Hash        string `json:"hash,omitempty"`
	Authentihash string `json:"authentihash,omitempty"`
	CDHash      string `json:"cdhash,omitempty"`
	MD5         string `json:"md5,omitempty"`
	SHA1        string `json:"sha1,omitempty"`
	SHA224      string `json:"sha224,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	SHA384      string `json:"sha384,omitempty"`
	SHA3224     string `json:"sha3224,omitempty"`
	SHA3256     string `json:"sha3256,omitempty"`
	SHA3384     string `json:"sha3384,omitempty"`
	SHA3512     string `json:"sha3512,omitempty"`
	SHA512      string `json:"sha512,omitempty"`
	SHA512224   string `json:"sha512224,omitempty"`
	SHA512256   string `json:"sha512256,omitempty"`
	Hex         string `json:"hex,omitempty"`
	Base64      string `json:"base64,omitempty"`

	// System
	OperatingSystem string `json:"operatingSystem,omitempty"`
	ChromeExtension string `json:"chromeExtension,omitempty"`
	MobileAppID     string `json:"mobileAppId,omitempty"`

	// Vulnerability
	CPE string `json:"cpe,omitempty"`
	CVE string `json:"cve,omitempty"`

	// Malware
	Malware       string `json:"malware,omitempty"`
	MalwareFamily string `json:"malwareFamily,omitempty"`
	MalwareType   string `json:"malwareType,omitempty"`

	// PGP keys
	PGPPrivateKey string `json:"pgpPrivateKey,omitempty"`
	PGPPublicKey  string `json:"pgpPublicKey,omitempty"`

	// Resources
	Connections    int64      `json:"connections,omitempty"`
	UsedCPUPercent int        `json:"usedCpuPercent,omitempty"`
	UsedMemPercent int        `json:"usedMemPercent,omitempty"`
	TotalCPUUnits  int        `json:"totalCpuUnits,omitempty"`
	TotalMem       int64      `json:"totalMem,omitempty"`
	Disks          []DiskInfo `json:"disks,omitempty"`
}

type Impact struct {
	Score int    `json:"score,omitempty"`
	Label string `json:"label,omitempty"`
}

type AlertEvent struct {
	ID    string `json:"id,omitempty"`
	Index string `json:"index,omitempty"`
}

type UtmAlert struct {
	Timestamp         string          `json:"@timestamp,omitempty"`
	ID                string          `json:"id,omitempty"`
	ParentID          string          `json:"parentId,omitempty"`
	Status            int             `json:"status,omitempty"`
	StatusLabel       string          `json:"statusLabel,omitempty"`
	StatusObservation string          `json:"statusObservation,omitempty"`
	IsIncident        bool            `json:"isIncident,omitempty"`
	IncidentDetail    *IncidentDetail `json:"incidentDetail,omitempty"`
	Name              string          `json:"name,omitempty"`
	Category          string          `json:"category,omitempty"`
	Severity          int             `json:"severity,omitempty"`
	SeverityLabel     string          `json:"severityLabel,omitempty"`
	Description       string          `json:"description,omitempty"`
	Solution          string          `json:"solution,omitempty"`
	Technique         string          `json:"technique,omitempty"`
	Reference         []string        `json:"reference,omitempty"`
	DataType          string          `json:"dataType,omitempty"`
	Impact            *Impact         `json:"impact,omitempty"`
	ImpactScore       int             `json:"impactScore,omitempty"`
	DataSource        string          `json:"dataSource,omitempty"`
	Adversary         *Side           `json:"adversary,omitempty"`
	Target            *Side           `json:"target,omitempty"`
	Events            []AlertEvent    `json:"events,omitempty"`
	LastEvent         *AlertEvent     `json:"lastEvent,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	Notes             string          `json:"notes,omitempty"`
	TagRulesApplied   []int64         `json:"tagRulesApplied,omitempty"`
	DeduplicatedBy    []string        `json:"deduplicatedBy,omitempty"`
	Logs              []string        `json:"logs,omitempty"`
	History           []AlertHistory  `json:"history,omitempty"`
	AssetGroupName    string          `json:"assetGroupName,omitempty"`
	AssetGroupID      int64           `json:"assetGroupId,omitempty"`
}

// AlertHistory is a single change-history entry embedded in the alert document.
// It replaces the former Postgres utm_alert_log table: the change history now
// lives under the alert itself in OpenSearch. Named "history" (not "log") to
// avoid confusion with `Logs`, which holds the source log events that raised
// the alert. Kept deliberately lean (no full old-alert snapshot) to avoid
// recursive document bloat.
type AlertHistory struct {
	User      string `json:"user,omitempty"`
	Action    string `json:"action,omitempty"`
	Message   string `json:"message,omitempty"`
	NewValue  string `json:"newValue,omitempty"`  // compact JSON of the changed fields
	Timestamp string `json:"timestamp,omitempty"` // RFC3339
}

// AlertAction enumerates the change actions recorded in an alert's history.
type AlertAction string

const (
	ActionUpdateStatus   AlertAction = "UPDATE_STATUS"
	ActionUpdateTags     AlertAction = "UPDATE_TAGS"
	ActionUpdateNotes    AlertAction = "UPDATE_NOTES"
	ActionUpdateSolution AlertAction = "UPDATE_SOLUTION"
	ActionMarkAsIncident AlertAction = "MARK_AS_INCIDENT"
)

// AlertStatus is the lifecycle state of an alert.
type AlertStatus int

const (
	AlertStatusAutomaticReview AlertStatus = 1
	AlertStatusOpen            AlertStatus = 2
	AlertStatusInReview        AlertStatus = 3
	AlertStatusCompleted       AlertStatus = 5
)

func StatusName(s AlertStatus) string {
	switch s {
	case AlertStatusAutomaticReview:
		return "Automatic review"
	case AlertStatusOpen:
		return "Open"
	case AlertStatusInReview:
		return "In review"
	case AlertStatusCompleted:
		return "Completed"
	default:
		return ""
	}
}

func IsValid(s AlertStatus) bool {
	switch s {
	case AlertStatusAutomaticReview, AlertStatusOpen, AlertStatusInReview, AlertStatusCompleted:
		return true
	default:
		return false
	}
}
