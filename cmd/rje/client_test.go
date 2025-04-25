package main

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
)

// Just using contents of actual keys for ease, regenerate keys if deploying
const (
	// CA.pem
	VALID_CERT_AUTHORITY = `-----BEGIN CERTIFICATE-----
MIIBeDCCAR+gAwIBAgIUX0PSM84fikXPXS3ePXs6XxNdy1AwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNTdaFw0yNjA0MjQyMjQw
NTdaMBIxEDAOBgNVBAMMB1Jvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AASOnDwKYR9+xX2QhvzJm3LCDY8WeC6f6PXlrr9jc+q8EuF5zOpOBztG/yFrPQeM
VhnGcqcqYw05x4pVrOyFKmZjo1MwUTAdBgNVHQ4EFgQUXTbONPd2ubmK1N6BkVtd
oR6scRYwHwYDVR0jBBgwFoAUXTbONPd2ubmK1N6BkVtdoR6scRYwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiA/0CKtCAbqzvzVgRm131OpZrBYXVuB
apyEPQ2zCtMGeAIgSsAXC00zMBnyNWaH+T9R6xSZt7OSKL9Z08mwmERd+To=
-----END CERTIFICATE-----
`
	// server.crt
	VALID_SERVER_PUB = `-----BEGIN CERTIFICATE-----
MIIBpzCCAUygAwIBAgIUL54wEtZC3qx8VCLilpmZ1RQ9QwQwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNTdaFw0yNjA0MjQyMjQw
NTdaMBExDzANBgNVBAMMBnNlcnZlcjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BDkKCFoT95kqDCqaIlQAlh1tLDTFT/w+W5DaCWUDIXMq/c4Z+fAbE1BTnmb1gVOC
4v43fzlWFgFk8sFkFFTuk9WjgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1UdDgQWBBTm
XVy04kBZYnqGh05AB/Pn0k5d2zAfBgNVHSMEGDAWgBRdNs4093a5uYrU3oGRW12h
HqxxFjAKBggqhkjOPQQDAgNJADBGAiEA8lGSEY5S3Eb/eoPLJcuF5KvsasNPfmK+
dBk5YpzbISgCIQDw1XwF3cCNIUFYceqrRnSV8RmboYf4kcU4EbYH5c+dQw==
-----END CERTIFICATE-----
`
	// server.key
	VALID_SERVER_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIKWs3+S8mO7oJYTLNOegM92miGc7M2/7S1E6AKFWN19ZoAoGCCqGSM49
AwEHoUQDQgAEOQoIWhP3mSoMKpoiVACWHW0sNMVP/D5bkNoJZQMhcyr9zhn58BsT
UFOeZvWBU4Li/jd/OVYWAWTywWQUVO6T1Q==
-----END EC PRIVATE KEY-----
`
	// client.crt
	VALID_CLIENT_PUB = `-----BEGIN CERTIFICATE-----
MIIBrDCCAVKgAwIBAgIUL54wEtZC3qx8VCLilpmZ1RQ9QwUwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNTdaFw0yNjA0MjQyMjQw
NTdaMBcxFTATBgNVBAMMDHZhbGlkX2NsaWVudDBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABGeBdGlyRUqUl/3QxVkMiQA5FDIEN+IqC+CKbpep341cVMwrJ0x2Gao6
nhmsC67Dm7GeId4f+tEtcOjOBhBLGgajgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0l
BAwwCgYIKwYBBQUHAwIwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1Ud
DgQWBBQmc1AZOWQ4yAiqQr7S32cmcRU6njAfBgNVHSMEGDAWgBRdNs4093a5uYrU
3oGRW12hHqxxFjAKBggqhkjOPQQDAgNIADBFAiEA4KKu8m9pgGKkX9zd3WImk3Oo
Oz9IxD0BqYXUpLQds3UCIHS3p3MH3rx/gHLYPtoUObJLyJHXkLN1srfw22FCcMoK
-----END CERTIFICATE-----
`
	// client.key
	VALID_CLIENT_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAc43Z7+OGYRq46xV8j6IC9nWaiHqXLslrPEgUhu8++loAoGCCqGSM49
AwEHoUQDQgAEZ4F0aXJFSpSX/dDFWQyJADkUMgQ34ioL4Ipul6nfjVxUzCsnTHYZ
qjqeGawLrsObsZ4h3h/60S1w6M4GEEsaBg==
-----END EC PRIVATE KEY-----
`
	// CA_fail.pem
	INVALID_CERT_AUTHORITY = `-----BEGIN CERTIFICATE-----
MIIBlTCCATugAwIBAgIUUNTKPwlDXQbksOQTpKm63sLT6BwwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNTdaFw0yNjA0MjQyMjQw
NTdaMBIxEDAOBgNVBAMMB1Jvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AATLx50TBU3Yr+s6QhMB05kqCbNi+sKoCBIJd1E4vRJhVmldul7OexUr+Le/xHjE
XYdU1J0eb8E+1XWrJAeRQ9ZSo28wbTAdBgNVHQ4EFgQUoKPai9oZM2DSp8HUYszb
hH+hwxgwHwYDVR0jBBgwFoAUoKPai9oZM2DSp8HUYszbhH+hwxgwDwYDVR0TAQH/
BAUwAwEB/zAaBgNVHREEEzARgglsb2NhbGhvc3SHBAAAAAAwCgYIKoZIzj0EAwID
SAAwRQIhAPJuvMj5v1f/yMHc7GqTmahllIWmIM1a3vHFhjh8PriuAiBYTjTsdjST
CMtmxTqQbRrLDhrMUX7tTAJIIEwgd3Nn7g==
-----END CERTIFICATE-----
`
	// server_fail.crt
	INVALID_SERVER_PUB = `-----BEGIN CERTIFICATE-----
MIIBpjCCAUygAwIBAgIUUz/ze0Cu1D/o2gN//6L6U2a7lkUwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNThaFw0yNjA0MjQyMjQw
NThaMBExDzANBgNVBAMMBnNlcnZlcjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BG3tfFPvx+5pQhTL3u0szsX45vFAuOfpihYNqtD0ijfkbPuYLcES9r28d843JOKd
LYIVPZlzIa5e40RvPaSrALGjgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1UdDgQWBBSK
uFyiywirrZCuqtboDd6ZjxNrvzAfBgNVHSMEGDAWgBSgo9qL2hkzYNKnwdRizNuE
f6HDGDAKBggqhkjOPQQDAgNIADBFAiALblI3akDnn59wjYUx9hK+F3XBpnpjX4rs
yhIt6hAtyQIhALr1QWfECsWQM5w3OQRS1G1C9L8FLQdZcfQm8am7JATu
-----END CERTIFICATE-----
`
	// server_fail.key
	INVALID_SERVER_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOUheUUoU2pgxQbtDJZ/5Ep4ka3f4hOjOYkasG1STkCsoAoGCCqGSM49
AwEHoUQDQgAEbe18U+/H7mlCFMve7SzOxfjm8UC45+mKFg2q0PSKN+Rs+5gtwRL2
vbx3zjck4p0tghU9mXMhrl7jRG89pKsAsQ==
-----END EC PRIVATE KEY-----
`
	// client_fail.crt
	INVALID_CLIENT_PUB = `-----BEGIN CERTIFICATE-----
MIIBqDCCAU+gAwIBAgIUUz/ze0Cu1D/o2gN//6L6U2a7lkYwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjQyMjQwNThaFw0yNjA0MjQyMjQw
NThaMBQxEjAQBgNVBAMMCWZhaWxfdXNlcjBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABCo1zeCAwDillp95B7ZWAOTr/NwOD9yor1CQDbwxHcPDKlCxcHgM+qrvhtWs
jFyK4req5mxvogBowliXx4rtZvyjgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0lBAww
CgYIKwYBBQUHAwIwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1UdDgQW
BBSxF1g54ygnUGiG4L+8cP143Qe6mTAfBgNVHSMEGDAWgBSgo9qL2hkzYNKnwdRi
zNuEf6HDGDAKBggqhkjOPQQDAgNHADBEAiBSUpLZKyJP66kOFPHbJw+p+spDstAN
OgjFk8SWSEmgkwIgV0i/boauOYQdXJAGcieXNmceUtJ9Vgh7w5U9Wpz5Wlw=
-----END CERTIFICATE-----
`
	// client_fail.key
	INVALID_CLIENT_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIASCVutGYAKEhdFxi4k8dMcrP6SCCfX5+zXmPDBnAZMZoAoGCCqGSM49
AwEHoUQDQgAEKjXN4IDAOKWWn3kHtlYA5Ov83A4P3KivUJANvDEdw8MqULFweAz6
qu+G1ayMXIrit6rmbG+iAGjCWJfHiu1m/A==
-----END EC PRIVATE KEY-----
`
)

func startServer(certFile []byte, keyFile []byte, certAuthorityFile []byte) (func(), string, error) {
	grpcServer, listener, err := createGrpcServer(0, certFile, keyFile, certAuthorityFile)
	if err != nil {
		return nil, "", err
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return func() {
		grpcServer.GracefulStop()
		listener.Close()
	}, listener.Addr().String(), nil
}

func TestClientArgs(t *testing.T) {
	if _, _, err := splitFlagsAndRemoteCommand([]string{"start"}); err == nil {
		t.Error("no arguments should fail")
	}

	if _, _, err := splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:0"}); err == nil {
		t.Error("no remote command separator should fail")
	}

	if _, _, err := splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:0", "--"}); err == nil {
		t.Error("no remote command should fail")
	}

	if _, _, err := splitFlagsAndRemoteCommand([]string{"start", "-h"}); err == nil || !errors.Is(pflag.ErrHelp, err) {
		t.Error("help argument should be valid")
	}

	if args, remoteJob, err := splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:5555", "--", "ls"}); err != nil {
		t.Error(fmt.Sprintf("command should pass: %s", err.Error()))
	} else {
		if len(remoteJob) != 1 {
			t.Error("failed to parse remote job")
		}
		if args.server != "localhost:5555" {
			t.Error("server argument failed to parse")
		}
	}

	if args, remoteJob, err := splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:5555", "--", "ls", "-lsa"}); err != nil {
		t.Error(fmt.Sprintf("multiple arguments should pass%s", err.Error()))
	} else {
		if len(remoteJob) != 2 {
			t.Error("failed to parse remote job")
		}
		if args.server != "localhost:5555" {
			t.Error("server argument failed to parse")
		}
	}
}

func TestClientConnection(t *testing.T) {
	shutdownServer, connAddr, err := startServer([]byte(VALID_SERVER_PUB), []byte(VALID_SERVER_PRIV), []byte(VALID_CERT_AUTHORITY))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	validCommand := []string{"start", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "--", "ls"}
	failCommand := []string{"start", "-s", connAddr, "--ca-file", INVALID_CERT_AUTHORITY, "--cert-file", INVALID_CLIENT_PUB, "--key-file", INVALID_CLIENT_PRIV, "--", "ls"}

	if err = start(failCommand); err == nil {
		t.Error(fmt.Sprintf("Connection should fail with bad keys"))
	}

	if err = start(validCommand); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Connection should pass with good keys: %s", err.Error()))
	}
}

func TestClientCommands(t *testing.T) {
	shutdownServer, connAddr, err := startServer([]byte(VALID_SERVER_PUB), []byte(VALID_SERVER_PRIV), []byte(VALID_CERT_AUTHORITY))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	if err := start([]string{"start", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "--", "ls"}); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Start should behave correctly: %s", err.Error()))
	}

	if err := status([]string{"status", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Status should behave correctly: %s", err.Error()))
	}

	if err := stop([]string{"stop", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Stop should behave correctly: %s", err.Error()))
	}

	if err := tail([]string{"tail", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Tail should behave correctly: %s", err.Error()))
	}
}

func TestClientAuthorized(t *testing.T) {
	shutdownServer, connAddr, err := startServer([]byte(VALID_SERVER_PUB), []byte(VALID_SERVER_PRIV), []byte(VALID_CERT_AUTHORITY))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	if err := start([]string{"start", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "--", "ls"}); err != nil && grpc_status.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Start should behave correctly: %s", err.Error()))
	}

	if err := start([]string{"start", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "--", "foobar"}); !errors.Is(ErrUnAuth, err) {
		t.Error(fmt.Sprintf("User shouldn't have access to foobar: %s", err.Error()))
	}
}

func TestClientUnAuthorized(t *testing.T) {
	shutdownServer, connAddr, err := startServer([]byte(INVALID_SERVER_PUB), []byte(INVALID_SERVER_PRIV), []byte(INVALID_CERT_AUTHORITY))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	if err := start([]string{"start", "-s", connAddr, "--ca-file", INVALID_CERT_AUTHORITY, "--cert-file", INVALID_CLIENT_PUB, "--key-file", INVALID_CLIENT_PRIV, "--", "ls"}); !errors.Is(ErrUnAuth, err) {
		t.Error(fmt.Sprintf("User shouldn't have access to ls: %s", err.Error()))
	}

	if err := start([]string{"start", "-s", connAddr, "--ca-file", INVALID_CERT_AUTHORITY, "--cert-file", INVALID_CLIENT_PUB, "--key-file", INVALID_CLIENT_PRIV, "--", "cat"}); !errors.Is(ErrUnAuth, err) {
		t.Error(fmt.Sprintf("User shouldn't have access to cat: %s", err.Error()))
	}
}
