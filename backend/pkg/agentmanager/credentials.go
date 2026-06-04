package agentmanager

import "context"

type internalKeyCreds struct {
	key string
}

func (c *internalKeyCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"internal-key": c.key}, nil
}

func (c *internalKeyCreds) RequireTransportSecurity() bool {
	return false
}
