package common_models

type AssetSensitivity struct {
	Name            string
	Hostnames       []string
	Ips             []string
	Confidentiality int
	Integrity       int
	Availability    int
}
