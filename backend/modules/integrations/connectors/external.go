package connectors

import "context"

type CredentialVerifier interface {
	Verify(module string, config map[string]string) error
}

type Cipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

// TenantFileToggler enables/disables a module's OWN tenant config file (the
// `.disabled` suffix) on activate/deactivate, so the file-reading plugins
// stop/start. Backed by repository.TenantStore. This is the integration's own
// file — not a cross-module side effect; other modules instead pull state via
// the is-active endpoint.
type TenantFileToggler interface {
	SetActiveByModule(ctx context.Context, moduleName string, active bool) error
}
