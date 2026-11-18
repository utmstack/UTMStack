package signatures

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
)

// These are Authenticode/CRB purposes, not general X.509 authorization. The
// pinned scanner processes KU/BC in wire order, unions EKU purposes, ignores
// other extension payloads (including critical constraints), and supplies
// defaults only when an extensions container is present. Keep this policy
// local; the original TBS must still authenticate against a trusted issuer.
type authenticodeCertificatePolicy struct {
	certSign, codeSign, timeSign bool
}

func certificateExtensionPolicy(fields []asn1.RawValue, algorithm int) (authenticodeCertificatePolicy, error) {
	var policy authenticodeCertificatePolicy
	end := algorithm + 5
	if len(fields) < end {
		return policy, errRejectedAuthenticodeCertificate
	}
	if algorithm == 1 {
		if len(fields) != end {
			return policy, errRejectedAuthenticodeCertificate
		}
		policy.certSign, policy.timeSign = true, true
		return policy, nil
	}
	last, work := 0, 0
	for position, field := range fields[end:] {
		if field.Class != asn1.ClassContextSpecific || !field.IsCompound || field.Tag <= last || field.Tag > 3 {
			return policy, errRejectedAuthenticodeCertificate
		}
		last = field.Tag
		if field.Tag != 3 {
			continue // The reference does not interpret issuer/subject unique IDs.
		}
		exts, tail, err := certificateKeyObject(field.Bytes, 0x30)
		if err != nil {
			return policy, errRejectedAuthenticodeCertificate
		}
		if len(tail) != 0 {
			// The pinned reader stops the optional-field loop here without
			// interpreting extensions or applying their defaults. A remaining
			// TBS field still makes the certificate invalid.
			if position != len(fields[end:])-1 {
				return policy, errRejectedAuthenticodeCertificate
			}
			return policy, nil
		}
		haveKU, haveEKU := false, false
		for len(exts) > 0 {
			work++
			if work > 131072 {
				return policy, errInvalidAuthenticode
			}
			ext, rest, err := certificateKeyObject(exts, 0x30)
			if err != nil {
				return policy, err
			}
			exts = rest
			oid, rest, err := certificateKeyObject(ext, 6)
			if err != nil {
				return policy, err
			}
			if len(rest) > 0 && rest[0] == 1 {
				critical, remaining, err := certificateKeyObject(rest, 1)
				if err != nil || len(critical) != 1 {
					return policy, errRejectedAuthenticodeCertificate
				}
				rest = remaining
			}
			value, tail, err := certificateKeyObject(rest, 4)
			if err != nil || len(tail) != 0 {
				return policy, errRejectedAuthenticodeCertificate
			}
			if len(oid) != 3 || oid[0] != 0x55 || oid[1] != 0x1d {
				continue
			}
			switch oid[2] {
			case 15: // KeyUsage
				haveKU = true
				if len(value) < 4 || len(value) > 5 || value[0] != 3 || int(value[1]) != len(value)-2 || value[2] > 7 {
					return policy, errRejectedAuthenticodeCertificate
				}
				usage := value[3]
				if len(value) == 4 {
					usage &^= byte((1 << value[2]) - 1)
				}
				policy.certSign = usage&4 != 0
			case 37: // ExtendedKeyUsage
				haveEKU = true
				purposes, tail, err := certificateKeyObject(value, 0x30)
				if err != nil || len(tail) != 0 {
					return policy, errRejectedAuthenticodeCertificate
				}
				for len(purposes) > 0 {
					work++
					if work > 131072 {
						return policy, errInvalidAuthenticode
					}
					purpose, rest, err := certificateKeyObject(purposes, 6)
					if err != nil {
						return policy, err
					}
					purposes = rest
					if len(purpose) == 8 || len(purpose) == 10 {
						policy.codeSign = policy.codeSign || bytes.Equal(purpose[:8], []byte{0x2b, 6, 1, 5, 5, 7, 3, 3})
						policy.timeSign = policy.timeSign || bytes.Equal(purpose[:8], []byte{0x2b, 6, 1, 5, 5, 7, 3, 8})
					}
				}
			case 19: // BasicConstraints: only its first child determines cA.
				constraint, _, err := certificateKeyObject(value, 0x30)
				if err != nil {
					return policy, err
				}
				policy.certSign = false
				if len(constraint) > 0 {
					switch constraint[0] {
					case 1:
						flag, _, err := certificateKeyObject(constraint, 1)
						if err != nil || len(flag) != 1 {
							return policy, errRejectedAuthenticodeCertificate
						}
						policy.certSign = flag[0] != 0
					case 2:
						if _, _, err := certificateKeyObject(constraint, 2); err != nil {
							return policy, err
						}
					default:
						return policy, errRejectedAuthenticodeCertificate
					}
				}
			}
		}
		if !haveKU {
			policy.certSign = true
		}
		if !haveEKU {
			policy.codeSign, policy.timeSign = true, true
		}
	}
	return policy, nil
}

func certificatePolicyFromTBS(data []byte) (authenticodeCertificatePolicy, error) {
	var empty authenticodeCertificatePolicy
	tbs, tail, err := authObject(data, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 || len(data) > 8<<20 {
		return empty, errInvalidAuthenticode
	}
	var fields []asn1.RawValue
	for rest := tbs.Bytes; len(rest) > 0; {
		if len(fields) >= 16 {
			return empty, errInvalidAuthenticode
		}
		field, next, err := asn1Object(rest)
		if err != nil {
			return empty, err
		}
		fields = append(fields, field)
		rest = next
	}
	algorithm := 1
	if len(fields) > 0 && fields[0].Class == asn1.ClassContextSpecific {
		algorithm = 2
	}
	return certificateExtensionPolicy(fields, algorithm)
}

func (p authenticodeCertificatePolicy) applyMetadata(cert *x509.Certificate) {
	// The metadata copy represents this reader's purposes, not PKIX constraints.
	cert.BasicConstraintsValid, cert.IsCA = true, p.certSign
	cert.MaxPathLen, cert.MaxPathLenZero = -1, false
	cert.KeyUsage = x509.KeyUsageDigitalSignature
	if p.certSign {
		cert.KeyUsage |= x509.KeyUsageCertSign
	}
	if p.codeSign {
		cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageCodeSigning)
	}
	if p.timeSign {
		cert.ExtKeyUsage = append(cert.ExtKeyUsage, x509.ExtKeyUsageTimeStamping)
	}
}
