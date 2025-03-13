package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadHTTPTLSCredentials(crtName string) (*tls.Config, error) {
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

func LoadGRPCTLSCredentials(crtName string) (credentials.TransportCredentials, error) {
	config, err := LoadHTTPTLSCredentials(crtName)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(config), nil
}
