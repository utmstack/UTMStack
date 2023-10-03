package processor

type MicrosoftLoginResponse struct {
	TokenType   string `json:"token_type"`
	Expires     int    `json:"expires_in"`
	ExtExpires  int    `json:"ext_expires_in"`
	AccessToken string `json:"access_token"`
}

type StartSubscriptionResponse struct {
	ContentType string      `json:"contentType"`
	Status      string      `json:"status"`
	WebHook     interface{} `json:"webhook"`
}

type ContentList struct {
	ContentUri        string `json:"contentUri"`
	ContentId         string `json:"contentId"`
	ContentType       string `json:"contentType"`
	ContentCreated    string `json:"contentCreated"`
	ContentExpiration string `json:"contentExpiration"`
}

type ContentDetailsResponse []map[string]interface{}
