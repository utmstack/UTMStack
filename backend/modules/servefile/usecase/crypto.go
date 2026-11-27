package usecase

import (
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

// fileCipher stores binary payloads through the repo-wide pkg/secret scheme.
type fileCipher = secret.BytesCipher

// Same constants as pkg/secret so ciphertexts stay portable.
var newFileCipher = secret.NewBytesCipher
