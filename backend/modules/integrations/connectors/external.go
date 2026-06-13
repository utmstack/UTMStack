package connectors

import "context"

type CredentialVerifier interface {
	Verify(module string, config map[string]string) error
}

type Cipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}
type TenantFileToggler interface {
	SetActiveByModule(ctx context.Context, moduleName string, active bool) error
}
