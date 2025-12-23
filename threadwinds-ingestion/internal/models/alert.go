package models

type Alert struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Timestamp   string `json:"@timestamp"`
	DataType    string `json:"dataType"`
	DataSource  string `json:"dataSource"`
	Severity    int    `json:"severity"`
	Status      int    `json:"status"`
	Source      *Host  `json:"source,omitempty"`
	Destination *Host  `json:"destination,omitempty"`
	ASO         string `json:"aso,omitempty"`
	ASN         int    `json:"asn,omitempty"`
}

type Host struct {
	IP                  string    `json:"ip,omitempty"`
	Host                string    `json:"host,omitempty"`
	User                string    `json:"user,omitempty"`
	Port                int       `json:"port,omitempty"`
	City                string    `json:"city,omitempty"`
	Country             string    `json:"country,omitempty"`
	Coordinates         []float64 `json:"coordinates,omitempty"`
	ASO                 string    `json:"aso,omitempty"`
	ASN                 int       `json:"asn,omitempty"`
	AccuracyRadius      int       `json:"accuracyRadius,omitempty"`
	IsAnonymousProxy    bool      `json:"isAnonymousProxy,omitempty"`
	IsSatelliteProvider bool      `json:"isSatelliteProvider,omitempty"`
}
