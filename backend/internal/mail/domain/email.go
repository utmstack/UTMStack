package domain

type EmailConfig struct {
	Host     string
	Port     string
	Password string
	Username string
	Orgname  string
	SmtpAuth string

	ProtocolValue string
	PortTlsValue  string
	PortSslValue  string
	PortNoneValue string
	From          string
	BaseUrl       string
}
