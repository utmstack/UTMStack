package domain

type MailConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"` // SMTP submission ports (587/465/25) exceed uint8
	AuthType string `json:"authType"`
	From     string `json:"from"`
}
