package socai

type Module struct {
	client *SocAIClient
}

func NewModule(baseURL, internalKey string) *Module {
	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
	}
}
