package certs

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"

	"google.golang.org/grpc/credentials"
)

//go:embed *
var CertsFS embed.FS

func LoadCerts(certFile string, keyFile string, certAuthorityFile string, asClient bool) (credentials.TransportCredentials, error) {
	// Load files from embed
	certPEMBlock, err := CertsFS.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	keyPEMBlock, err := CertsFS.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	trustedCert, err := CertsFS.ReadFile(certAuthorityFile)
	if err != nil {
		return nil, err
	}

	// Load server certificates
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	// Put the CA certificate into the certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		return nil, errors.New("Failed to append trusted certificate to certificate pool")
	}

	// Create the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}

	if asClient {
		tlsConfig.ClientCAs = certPool
	}

	// Return new TLS credentials based on the TLS configuration
	return credentials.NewTLS(tlsConfig), nil
}
