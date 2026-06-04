package domain

type MailConfig struct {
	Host     string
	Username string
	Password string
	Port     uint8
	AuthType string
	From     string
}
