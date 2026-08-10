package dto

type CollectorResponse struct {
	ID       int32  `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Version  string `json:"version"`
	LastSeen string `json:"lastSeen,omitempty"`
	Status   string `json:"status"`
}

type SetDataTypeConfigRequest struct {
	Enabled         *bool  `json:"enabled" binding:"required"`
	Proto           string `json:"proto" binding:"required"`
	Port            string `json:"port,omitempty"`
	TLS             *bool  `json:"tls,omitempty"`
	Auth            string `json:"auth,omitempty"`
	Path            string `json:"path,omitempty"`
	SignatureHeader string `json:"signatureHeader,omitempty"`
}

type ConfigKnowledgeResponse struct {
	Accepted        bool   `json:"accepted"`
	RequestID       string `json:"requestId,omitempty"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
	GeneratedSecret string `json:"generatedSecret,omitempty"`
}

type SetForwarderCertificatesRequest struct {
	CertPem *string `json:"certPem" binding:"required"`
	KeyPem  *string `json:"keyPem" binding:"required"`
	CaPem   *string `json:"caPem,omitempty"`
}

type TLSStatusResponse struct {
	Available  bool   `json:"available"`
	CertExists bool   `json:"certExists"`
	KeyExists  bool   `json:"keyExists"`
	CAExists   bool   `json:"caExists"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}

type GetDataTypeConfigResponse struct {
	Configured      bool   `json:"configured"`
	Enabled         *bool  `json:"enabled,omitempty"`
	Proto           string `json:"proto,omitempty"`
	Port            string `json:"port,omitempty"`
	TLS             *bool  `json:"tls,omitempty"`
	Auth            string `json:"auth,omitempty"`
	Path            string `json:"path,omitempty"`
	SignatureHeader string `json:"signatureHeader,omitempty"`
	ConfigStatus    string `json:"configStatus,omitempty"`
	LastError       string `json:"lastError,omitempty"`
}
