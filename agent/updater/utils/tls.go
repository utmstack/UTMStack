package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func LoadTLSCredentials(crtName string) (*tls.Config, error) {
	// Load the server's certificate
	serverCert, err := os.ReadFile(crtName)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(serverCert) {
		return nil, fmt.Errorf("failed to add server certificate to the certificate pool")
	}

	config := &tls.Config{
		RootCAs: certPool,
	}

	return config, nil
}
