package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

type certificateFileContents struct {
	certificateFileContents   []byte
	keyFileContents           []byte
	certAuthorityFileContents []byte
}

func readCertsFromFiles(certFile string, keyFile string, certAuthorityFile string) (*certificateFileContents, error) {
	certFileContent, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	keyFileContent, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	certAuthorityFileContent, err := os.ReadFile(certAuthorityFile)
	if err != nil {
		return nil, err
	}

	return &certificateFileContents{
		certificateFileContents:   certFileContent,
		keyFileContents:           keyFileContent,
		certAuthorityFileContents: certAuthorityFileContent,
	}, nil
}

func loadCerts(certFile []byte, keyFile []byte, certAuthorityFile []byte) (credentials.TransportCredentials, string, error) {
	// Load server certificates
	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, "", err
	}

	// Put the CA certificate into the certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certAuthorityFile) {
		return nil, "", errors.New("Failed to append trusted certificate to certificate pool")
	}

	// Create the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Return new TLS credentials based on the TLS configuration
	tls := credentials.NewTLS(tlsConfig)

	return tls, cert.Leaf.Subject.CommonName, nil
}
