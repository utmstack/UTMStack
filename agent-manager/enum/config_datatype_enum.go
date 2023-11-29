package enum

type UTMConfigDataType string

const (
	TextType     UTMConfigDataType = "text"
	TelType                        = "tel"
	PasswordType                   = "password"
	EmailType                      = "email"
	NumberType                     = "number"
	BooleanType                    = "bool"
	ListType                       = "list"
)
