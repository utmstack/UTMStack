package signatures

import (
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"time"
)

// This verifier is local to CRB-authenticated Authenticode. It does not
// change Go's TLS/X.509 policy or make historical algorithms trust anchors.
func authenticodeCertificateHash(algorithm x509.SignatureAlgorithm) crypto.Hash {
	switch algorithm {
	case x509.MD5WithRSA:
		return crypto.MD5
	case x509.SHA1WithRSA:
		return crypto.SHA1
	case x509.SHA256WithRSA:
		return crypto.SHA256
	case x509.SHA384WithRSA:
		return crypto.SHA384
	case x509.SHA512WithRSA:
		return crypto.SHA512
	default:
		return 0
	}
}

func verifyAuthenticodeCertificateSignature(cert *x509.Certificate, key any) bool {
	valid, _ := authenticodeCertificateProof(cert, key)
	return valid
}

// These helpers expose the same ordered, bounded trust resolution used by
// PE, catalog and timestamp verification. They do not enumerate alternate paths
// independently for each requested purpose.
func (d *Database) historicalAuthenticodeChain(signer *x509.Certificate, certs []*x509.Certificate, at time.Time, timeSigning bool) []*x509.Certificate {
	chain, _ := d.authenticodeChainChecked(signer, certs, at, timeSigning)
	return chain
}

func (d *Database) authenticodeChainChecked(signer *x509.Certificate, certs []*x509.Certificate, at time.Time, timeSigning bool) ([]*x509.Certificate, error) {
	if signer == nil {
		return nil, nil
	}
	store, err := d.resolveAuthenticodeStore(certs, false)
	if err != nil {
		return nil, err
	}
	fingerprint := sha256.Sum256(signer.Raw)
	for _, entry := range store.entries {
		if entry.cert == nil || sha256.Sum256(entry.cert.Raw) != fingerprint {
			continue
		}
		if timeSigning && !entry.identity.timeSign || !timeSigning && !entry.identity.codeSign {
			continue
		}
		if !timeSigning && (at.Before(entry.cert.NotBefore) || at.After(entry.cert.NotAfter)) {
			continue
		}
		return entry.path, nil
	}
	if store.uncertain {
		return nil, errInvalidAuthenticode
	}
	return nil, nil
}
