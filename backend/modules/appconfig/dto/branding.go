package dto

import "time"

type BrandingResponse struct {
	Enabled        bool       `json:"enabled"`
	ProductName    string     `json:"productName"`
	LogoURL        string     `json:"logoUrl"`
	LogoDarkURL    string     `json:"logoDarkUrl"`
	FaviconURL     string     `json:"faviconUrl"`
	ReportLogoURL  string     `json:"reportLogoUrl"`  // logo embedded in report PDFs (legacy REPORT)
	ReportCoverURL string     `json:"reportCoverUrl"` // report PDF cover image (legacy REPORT_COVER)
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy      string     `json:"updatedBy,omitempty"`
}

type BrandingRequest struct {
	Enabled        bool   `json:"enabled"`
	ProductName    string `json:"productName"`
	LogoURL        string `json:"logoUrl"`
	LogoDarkURL    string `json:"logoDarkUrl"`
	FaviconURL     string `json:"faviconUrl"`
	ReportLogoURL  string `json:"reportLogoUrl"`  // logo embedded in report PDFs (legacy REPORT)
	ReportCoverURL string `json:"reportCoverUrl"` // report PDF cover image (legacy REPORT_COVER)
}

// BrandingPublic is the unauthenticated login-page view: identity assets only
// (report images are admin/server-side, not exposed here).
type BrandingPublic struct {
	Enabled     bool   `json:"enabled"`
	ProductName string `json:"productName"`
	LogoURL     string `json:"logoUrl"`
	LogoDarkURL string `json:"logoDarkUrl"`
	FaviconURL  string `json:"faviconUrl"`
}
