package socai

type Module struct {
	client *SocAIClient
}

func NewModule(baseURL, internalKey string) *Module {
	return &Module{
		client: NewSocAIClient(baseURL, internalKey),
	}
}

// Client returns the underlying SOC AI HTTP client. MCP tools call this to
// forward alert payloads to the external SOC AI service.
func (m *Module) Client() *SocAIClient { return m.client }
