package utils

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/shared/fs"
)

const (
	CertFilePermissions = 0644
	KeyFilePermissions  = 0600
	MinTLSVersion       = tls.VersionTLS12
	MaxTLSVersion       = tls.VersionTLS13
)

var certFilesMu sync.RWMutex

type TLSStatus struct {
	Available  bool   `json:"available"`
	CertExists bool   `json:"cert_exists"`
	KeyExists  bool   `json:"key_exists"`
	CAExists   bool   `json:"ca_exists"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}

type CertificateFiles struct {
	CertPath string
	KeyPath  string
	CAPath   string
}

func LoadIntegrationTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return nil, fmt.Errorf("error loading TLS certificate: %w", err)
	}

	return &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			certFilesMu.RLock()
			defer certFilesMu.RUnlock()
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, fmt.Errorf("error loading TLS certificate: %w", err)
			}
			return &cert, nil
		},
		MinVersion: MinTLSVersion,
		MaxVersion: MaxTLSVersion,
		CipherSuites: []uint16{
			// TLS 1.2 secure cipher suites - RSA key exchange
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// TLS 1.2 secure cipher suites - ECDSA key exchange (for ECDSA certificates)
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,    // Modern and fast
			tls.CurveP256, // NIST P-256
			tls.CurveP384, // NIST P-384
			tls.CurveP521, // NIST P-521
		},
		PreferServerCipherSuites: true,
	}, nil
}

func GetTLSStatus(certPath, keyPath, caPath string) TLSStatus {
	result := TLSStatus{
		CertExists: fs.Exists(certPath),
		KeyExists:  fs.Exists(keyPath),
		CAExists:   fs.Exists(caPath),
	}
	result.Available = result.CertExists && result.KeyExists

	if result.Available {
		if err := ValidateIntegrationCertificates(certPath, keyPath); err != nil {
			result.Error = err.Error()
		} else {
			result.Valid = true
		}
	}
	return result
}

func ValidateIntegrationCertificates(certPath, keyPath string) error {
	if !fs.Exists(certPath) {
		return fmt.Errorf("certificate file not found: %s", certPath)
	}

	if !fs.Exists(keyPath) {
		return fmt.Errorf("private key file not found: %s", keyPath)
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("invalid certificate or private key: %w", err)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("error parsing certificate: %w", err)
	}

	// 1. Check validity dates
	now := time.Now()
	if now.Before(x509Cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from: %s)",
			x509Cert.NotBefore.Format("2006-01-02 15:04:05 UTC"))
	}

	if now.After(x509Cert.NotAfter) {
		return fmt.Errorf("certificate has expired (valid until: %s)",
			x509Cert.NotAfter.Format("2006-01-02 15:04:05 UTC"))
	}

	// 2. Warn if the certificate expires soon (30 days)
	if now.Add(30 * 24 * time.Hour).After(x509Cert.NotAfter) {
		fmt.Printf("WARNING: Certificate expires soon (%s)\n",
			x509Cert.NotAfter.Format("2006-01-02 15:04:05 UTC"))
	}

	// 3. Check signature algorithm (reject weak algorithms)
	switch x509Cert.SignatureAlgorithm {
	case x509.SHA1WithRSA, x509.MD5WithRSA:
		return fmt.Errorf("certificate uses weak signature algorithm: %s (use SHA256+ instead)",
			x509Cert.SignatureAlgorithm)
	}

	// 4. Check RSA key size (minimum 2048 bits)
	if x509Cert.PublicKeyAlgorithm == x509.RSA {
		if rsaKey, ok := x509Cert.PublicKey.(*rsa.PublicKey); ok {
			keySize := rsaKey.Size() * 8 // Convert bytes to bits
			if keySize < 2048 {
				return fmt.Errorf("RSA key size too small: %d bits (minimum 2048 bits required)", keySize)
			}
		}
	}

	return nil
}

func LoadUserCertificatesWithStruct(src, dest CertificateFiles) error {
	// Validate source certificates
	if !fs.Exists(src.CertPath) {
		return fmt.Errorf("user certificate file not found: %s", src.CertPath)
	}
	if !fs.Exists(src.KeyPath) {
		return fmt.Errorf("user private key file not found: %s", src.KeyPath)
	}
	if err := ValidateIntegrationCertificates(src.CertPath, src.KeyPath); err != nil {
		return err
	}

	// Prepare destination directory
	certsDir := filepath.Dir(dest.CertPath)
	if err := fs.CreateDirIfNotExist(certsDir); err != nil {
		return fmt.Errorf("error creating certificates directory: %w", err)
	}

	certFilesMu.Lock()
	defer certFilesMu.Unlock()

	if err := copyFile(src.CertPath, dest.CertPath, CertFilePermissions); err != nil {
		return fmt.Errorf("error copying certificate: %w", err)
	}
	if err := copyFile(src.KeyPath, dest.KeyPath, KeyFilePermissions); err != nil {
		return fmt.Errorf("error copying private key: %w", err)
	}

	// Copy CA certificate (use source CA if exists, otherwise use cert as CA)
	caSource := src.CAPath
	if caSource == "" || !fs.Exists(caSource) {
		caSource = src.CertPath
	}
	if err := copyFile(caSource, dest.CAPath, CertFilePermissions); err != nil {
		return fmt.Errorf("error copying CA certificate: %w", err)
	}

	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".copyfile-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", dst, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once the rename below succeeds

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("set permissions on temp file for %s: %w", dst, err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file for %s: %w", dst, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", dst, err)
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename temp file to %s: %w", dst, err)
	}
	return nil
}
