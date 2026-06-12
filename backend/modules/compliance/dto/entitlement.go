package dto

type EntitlementRequest struct {
	Enterprise bool     `json:"enterprise"`
	Frameworks []string `json:"frameworks"`
	Controls   []string `json:"controls"`
}
