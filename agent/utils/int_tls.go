package utils

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	CertFilePermissions = 0644
	KeyFilePermissions  = 0600
	MinTLSVersion       = tls.VersionTLS12
	MaxTLSVersion       = tls.VersionTLS13
)

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
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading TLS certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   MinTLSVersion,
		MaxVersion:   MaxTLSVersion,
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

func ValidateIntegrationCertificates(certPath, keyPath string) error {
	if !CheckIfPathExist(certPath) {
		return fmt.Errorf("certificate file not found: %s", certPath)
	}

	if !CheckIfPathExist(keyPath) {
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
	if !CheckIfPathExist(src.CertPath) {
		return fmt.Errorf("user certificate file not found: %s", src.CertPath)
	}
	if !CheckIfPathExist(src.KeyPath) {
		return fmt.Errorf("user private key file not found: %s", src.KeyPath)
	}
	if err := ValidateIntegrationCertificates(src.CertPath, src.KeyPath); err != nil {
		return err
	}

	// Prepare destination directory
	certsDir := filepath.Dir(dest.CertPath)
	if err := CreatePathIfNotExist(certsDir); err != nil {
		return fmt.Errorf("error creating certificates directory: %w", err)
	}

	// Copy certificate files
	if err := copyFile(src.CertPath, dest.CertPath); err != nil {
		return fmt.Errorf("error copying certificate: %w", err)
	}
	if err := copyFile(src.KeyPath, dest.KeyPath); err != nil {
		return fmt.Errorf("error copying private key: %w", err)
	}

	// Copy CA certificate (use source CA if exists, otherwise use cert as CA)
	caSource := src.CAPath
	if caSource == "" || !CheckIfPathExist(caSource) {
		caSource = src.CertPath
	}
	if err := copyFile(caSource, dest.CAPath); err != nil {
		return fmt.Errorf("error copying CA certificate: %w", err)
	}

	// Set file permissions
	if err := os.Chmod(dest.CertPath, CertFilePermissions); err != nil {
		return fmt.Errorf("error setting certificate permissions: %w", err)
	}
	if err := os.Chmod(dest.KeyPath, KeyFilePermissions); err != nil {
		return fmt.Errorf("error setting private key permissions: %w", err)
	}
	if err := os.Chmod(dest.CAPath, CertFilePermissions); err != nil {
		return fmt.Errorf("error setting CA permissions: %w", err)
	}

	return nil
}
