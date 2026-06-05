package verifier

import (
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
)

type BackendVerifier struct{}

func NewBackendVerifier() *BackendVerifier { return &BackendVerifier{} }

var _ connectors.CredentialVerifier = (*BackendVerifier)(nil)

func (BackendVerifier) Verify(module string, c map[string]string) error {
	switch module {
	case "GCP":
		return verifyGCP(c)
	case "AWS_IAM_USER":
		return verifyAWSIAMUser(c)
	case "AZURE":
		return verifyAzure(c)
	case "O365":
		return verifyO365(c)
	case "BITDEFENDER":
		return verifyBitdefender(c)
	case "CROWDSTRIKE":
		return verifyCrowdStrike(c)
	default:
		return nil
	}
}

func required(c map[string]string, module string, keys ...string) error {
	for _, k := range keys {
		if strings.TrimSpace(c[k]) == "" {
			return fmt.Errorf("%s: %q is required", module, k)
		}
	}
	return nil
}
