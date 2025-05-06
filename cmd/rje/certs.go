package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"google.golang.org/grpc/credentials"
)

// Datatype to wrap the 3 required certificate files
type certificateFileContents struct {
	certificateFileContents   []byte
	keyFileContents           []byte
	certAuthorityFileContents []byte
}

// Loads the passed in certificate file contents in to a credentials.TransportCredentials type usable by the GRPC server
func loadCerts(certFile []byte, keyFile []byte, certAuthorityFile []byte) (credentials.TransportCredentials, error) {
	// Load server certificates
	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Put the CA certificate into the certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certAuthorityFile) {
		return nil, errors.New("Failed to append trusted certificate to certificate pool")
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

	return tls, nil
}
