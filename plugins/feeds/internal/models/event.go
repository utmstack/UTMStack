package models

type Event struct {
	ID     string `json:"id"`
	Origin *Side  `json:"origin,omitempty"`
	Target *Side  `json:"target,omitempty"`
}

type Side struct {
	// Network identification attributes
	IP          string       `json:"ip,omitempty"`
	Port        int          `json:"port,omitempty"`
	Host        string       `json:"host,omitempty"`
	Domain      string       `json:"domain,omitempty"`
	Mac         string       `json:"mac,omitempty"`
	Geolocation *Geolocation `json:"geolocation,omitempty"`
	URL         string       `json:"url,omitempty"`
	Cidr        string       `json:"cidr,omitempty"`

	// Certificate and fingerprint attributes
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	Ja3Fingerprint         string `json:"ja3Fingerprint,omitempty"`
	JarmFingerprint        string `json:"jarmFingerprint,omitempty"`
	SshBanner              string `json:"sshBanner,omitempty"`
	SshFingerprint         string `json:"sshFingerprint,omitempty"`

	// Web attributes
	Cookie   string `json:"cookie,omitempty"`
	JabberId string `json:"jabberId,omitempty"`

	// Email attributes
	EmailAddress     string `json:"emailAddress,omitempty"`
	EmailBody        string `json:"emailBody,omitempty"`
	EmailDisplayName string `json:"emailDisplayName,omitempty"`
	EmailSubject     string `json:"emailSubject,omitempty"`
	EmailThreadIndex string `json:"emailThreadIndex,omitempty"`
	EmailXMailer     string `json:"emailXMailer,omitempty"`
	Dkim             string `json:"dkim,omitempty"`
	DkimSignature    string `json:"dkimSignature,omitempty"`

	// WHOIS attributes
	WhoisRegistrant string `json:"whoisRegistrant,omitempty"`
	WhoisRegistrar  string `json:"whoisRegistrar,omitempty"`

	// Identity attributes
	User string `json:"user,omitempty"`

	// Process-related attributes
	Process                   string `json:"process,omitempty"`
	ProcessState              string `json:"processState,omitempty"`
	WindowsScheduledTask      string `json:"windowsScheduledTask,omitempty"`
	WindowsServiceDisplayName string `json:"windowsServiceDisplayName,omitempty"`
	WindowsServiceName        string `json:"windowsServiceName,omitempty"`

	// File-related attributes
	File        string `json:"file,omitempty"`
	Path        string `json:"path,omitempty"`
	Filename    string `json:"filename,omitempty"`
	SizeInBytes string `json:"sizeInBytes,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`

	// Hash-related attributes
	Authentihash string `json:"authentihash,omitempty"`
	Cdhash       string `json:"cdhash,omitempty"`
	MD5          string `json:"md5,omitempty"`
	SHA1         string `json:"sha1,omitempty"`
	SHA224       string `json:"sha224,omitempty"`
	SHA256       string `json:"sha256,omitempty"`
	SHA384       string `json:"sha384,omitempty"`
	SHA3224      string `json:"sha3224,omitempty"`
	SHA3256      string `json:"sha3256,omitempty"`
	SHA3384      string `json:"sha3384,omitempty"`
	SHA3512      string `json:"sha3512,omitempty"`
	SHA512       string `json:"sha512,omitempty"`
	SHA512224    string `json:"sha512224,omitempty"`
	SHA512256    string `json:"sha512256,omitempty"`
	Hex          string `json:"hex,omitempty"`
	Base64       string `json:"base64,omitempty"`

	// System-related attributes
	ChromeExtension string `json:"chromeExtension,omitempty"`
	MobileAppId     string `json:"mobileAppId,omitempty"`

	// Vulnerability-related attributes
	Cpe string `json:"cpe,omitempty"`
	Cve string `json:"cve,omitempty"`

	// Malware-related attributes
	Malware       string `json:"malware,omitempty"`
	MalwareFamily string `json:"malwareFamily,omitempty"`
	MalwareType   string `json:"malwareType,omitempty"`

	// Key-related attributes
	PgpPrivateKey string `json:"pgpPrivateKey,omitempty"`
	PgpPublicKey  string `json:"pgpPublicKey,omitempty"`
}

type FlattenedField struct {
	Path  string
	Key   string
	Value any
}

type Geolocation struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	ASN       int     `json:"asn,omitempty"`
	ASO       string  `json:"aso,omitempty"`
	Accuracy  int     `json:"accuracy,omitempty"`
}
