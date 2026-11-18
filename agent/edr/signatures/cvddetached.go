package signatures

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/smallstep/pkcs7"
)

// The PEM is distributed with ClamAV 1.5.4. Pinning its file digest prevents
// an accidental replacement of the embedded trust anchor.
//
//go:embed certs/clamav.crt
var officialClamAVRoot []byte

var officialClamAVRootSHA256 = [32]byte{
	0x60, 0x20, 0x0e, 0x4b, 0x4b, 0x1b, 0x1b, 0x72,
	0x57, 0xbb, 0x45, 0x50, 0x48, 0x7a, 0xeb, 0x14,
	0x20, 0x34, 0x34, 0x1f, 0x33, 0x00, 0xa8, 0x52,
	0x6e, 0x77, 0xa3, 0x13, 0x4a, 0x94, 0x0e, 0x9d,
}

// The detached-signature format was introduced at ClamAV feature level 230.
// It is independent of FunctionalityLevel, which selects detection rules.
const detachedSignatureLevel = 230

// VerifyOfficialDetachedCVD verifies a ClamAV .cvd.sign file over the exact
// container bytes. This is independent of the legacy CVD-header signature.
func VerifyOfficialDetachedCVD(cvd, sign []byte) error {
	if len(cvd) < 512 || !bytes.HasPrefix(cvd, []byte("ClamAV-VDB:")) {
		return errors.New("detached signature target is not a CVD")
	}
	rootDigest := sha256.Sum256(officialClamAVRoot)
	if subtle.ConstantTimeCompare(rootDigest[:], officialClamAVRootSHA256[:]) != 1 {
		return errors.New("embedded ClamAV trust root does not match its pin")
	}
	return verifyDetachedCVD(cvd, sign, officialClamAVRoot, detachedSignatureLevel, time.Now(), "ClamAV_datafiles_release")
}

func verifyDetachedCVD(cvd, sign, rootPEM []byte, level int, now time.Time, expectedSigner string) error {
	if len(sign) == 0 || len(sign) > 1<<20 || bytes.IndexByte(sign, 0) >= 0 {
		return errors.New("invalid detached signature file size or contents")
	}
	lines := bytes.Split(sign, []byte{'\n'})
	if len(lines) < 2 || string(bytes.TrimSuffix(lines[0], []byte{'\r'})) != "#clamsign-1.0" {
		return errors.New("unsupported ClamAV detached signature format")
	}
	block, rest := pem.Decode(rootPEM)
	if block == nil || block.Type != "CERTIFICATE" || len(bytes.TrimSpace(rest)) != 0 {
		return errors.New("invalid ClamAV trust root PEM")
	}
	root, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("invalid ClamAV trust root certificate")
	}
	roots := x509.NewCertPool()
	roots.AddCert(root)
	eligible := false
	for _, raw := range lines[1:] {
		line := strings.TrimSpace(string(raw))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) != 4 || fields[2] != "pkcs7-pem" || fields[0] == "" || fields[3] == "" {
			return errors.New("invalid ClamAV detached signature row")
		}
		minLevel, err := strconv.Atoi(fields[0])
		if err != nil || minLevel < 0 {
			return errors.New("invalid detached signature minimum feature level")
		}
		if fields[1] != "" {
			maxLevel, err := strconv.Atoi(fields[1])
			if err != nil || maxLevel < minLevel {
				return errors.New("invalid detached signature maximum feature level")
			}
			if level > maxLevel {
				continue
			}
		}
		if level < minLevel {
			continue
		}
		eligible = true
		der, err := base64.StdEncoding.Strict().DecodeString(fields[3])
		if err != nil || len(der) == 0 || len(der) > 1<<20 {
			return errors.New("invalid detached signature PKCS#7 encoding")
		}
		p7, err := pkcs7.Parse(der)
		if err != nil || len(p7.Content) != 0 || p7.GetOnlySigner() == nil {
			return errors.New("invalid detached PKCS#7 signed data")
		}
		if expectedSigner != "" && p7.GetOnlySigner().Subject.CommonName != expectedSigner {
			return errors.New("detached signature signer is not ClamAV's release identity")
		}
		p7.Content = cvd
		if err := p7.VerifyWithChainAtTime(roots, now); err != nil {
			return fmt.Errorf("ClamAV detached signature verification failed: %w", err)
		}
		return nil
	}
	if !eligible {
		return errors.New("no detached signature supports this engine feature level")
	}
	return errors.New("no valid ClamAV detached signature")
}
