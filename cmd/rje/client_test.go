package main

import (
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

// Just using contents of actual keys for ease, regenerate keys if deploying
const (
	VALID_CERT_AUTHORITY = `-----BEGIN CERTIFICATE-----
MIIBeDCCAR+gAwIBAgIUE+h2g5AUu73BD1YpvgpP1gl+cxMwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjIxNTQ2MDNaFw0yNjA0MjIxNTQ2
MDNaMBIxEDAOBgNVBAMMB1Jvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AATKuKji6qBz7+o73Dm1PHMBhCg98Mf6Ji1TNaUJCto4pT8c0Jd2porDtNn2nvft
n70PJk4SJcqj7pFaXYp6Hdq3o1MwUTAdBgNVHQ4EFgQUA/kmwC648hw0OfGEMrkr
BA6Z77swHwYDVR0jBBgwFoAUA/kmwC648hw0OfGEMrkrBA6Z77swDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiBmlkdhl4YUFiqquFT6SAwTCv/0gfjR
8OnpAisqELXfaAIgPJ57M4jmuv/ORDSxAPaoy+53/QsqLgb5rVGS3Tc0OGo=
-----END CERTIFICATE-----
`

	VALID_SERVER_PUB = `-----BEGIN CERTIFICATE-----
MIIBpzCCAUygAwIBAgIUL54wEtZC3qx8VCLilpmZ1RQ9QwIwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjIxNTQ2MDNaFw0yNjA0MjIxNTQ2
MDNaMBExDzANBgNVBAMMBnNlcnZlcjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BK5uQyH32WbgB7sWkIfR5YJJtBvcbDcGW/yxzkmzyZVSg77LZ1gmTKKoLW+K3r1L
b5QSrQsDiq/boxXagljZr5OjgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1UdDgQWBBRd
Ls2wZ3dOjkFGm3pep7kck/HdSjAfBgNVHSMEGDAWgBQD+SbALrjyHDQ58YQyuSsE
DpnvuzAKBggqhkjOPQQDAgNJADBGAiEAv9k1mRTsNx6XmQi9hKHCszVUDddBhLnK
Yt5HUrTYG8UCIQDGkB89pzlz5U6goAOeDWNZ0c1LAiIdFSSbSXPWnhH5Og==
-----END CERTIFICATE-----
`

	VALID_SERVER_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOPB5NUn8ryCI5cvawTp28mRomWBMii6i1ahPOG7zeKMoAoGCCqGSM49
AwEHoUQDQgAErm5DIffZZuAHuxaQh9Hlgkm0G9xsNwZb/LHOSbPJlVKDvstnWCZM
oqgtb4revUtvlBKtCwOKr9ujFdqCWNmvkw==
-----END EC PRIVATE KEY-----
`

	VALID_CLIENT_PUB = `-----BEGIN CERTIFICATE-----
MIIBrDCCAVKgAwIBAgIUL54wEtZC3qx8VCLilpmZ1RQ9QwMwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjIxNTQ2MDNaFw0yNjA0MjIxNTQ2
MDNaMBcxFTATBgNVBAMMDHZhbGlkX2NsaWVudDBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABD/FNHXN7SFTGcE13cqeMAOiJDC70Smr6FBhprFwUImDhPRfvIoaZ60e
f2aKwVBrpigWRheFPWzGDIjGxO+1ln2jgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0l
BAwwCgYIKwYBBQUHAwIwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1Ud
DgQWBBTXBdmprZy4kp65TMPaAGbq2Z3+kjAfBgNVHSMEGDAWgBQD+SbALrjyHDQ5
8YQyuSsEDpnvuzAKBggqhkjOPQQDAgNIADBFAiEA7xBJxSOX+W5ZueDLqjPmFt7c
ExcntRCkAb1vUI085a4CIH/IPxT8HvsKXNslGh0hUBRqjR26WdtXasqhfZiCJ8eK
-----END CERTIFICATE-----
`

	VALID_CLIENT_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIEch3ZLOJ2l2yI1KijKE8vmF36N88ZZw0FiJXvVrPXKToAoGCCqGSM49
AwEHoUQDQgAEP8U0dc3tIVMZwTXdyp4wA6IkMLvRKavoUGGmsXBQiYOE9F+8ihpn
rR5/ZorBUGumKBZGF4U9bMYMiMbE77WWfQ==
-----END EC PRIVATE KEY-----
`

	INVALID_CERT_AUTHORITY = `-----BEGIN CERTIFICATE-----
MIIBlTCCATugAwIBAgIUdvOkfnRhMThwMRuLbGz007yD664wCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjIxNTQ2MDNaFw0yNjA0MjIxNTQ2
MDNaMBIxEDAOBgNVBAMMB1Jvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AARC3EMxElcZCtz+mC4UFCkAzVOG1MOA1z4jxNk76L1gOWu4mrdb4KYgmRv3QRfl
ISU5RshaGY227IqPPnoZHxtFo28wbTAdBgNVHQ4EFgQUwvSvg5Ng6ImqSEyNVokB
R4Y0iTUwHwYDVR0jBBgwFoAUwvSvg5Ng6ImqSEyNVokBR4Y0iTUwDwYDVR0TAQH/
BAUwAwEB/zAaBgNVHREEEzARgglsb2NhbGhvc3SHBAAAAAAwCgYIKoZIzj0EAwID
SAAwRQIgdF1GFiq59GTN/Rldl7PvCmrESvmYIn8GISqex3C/XBsCIQChDBDWmZCh
m5WkKZfCObOG92xTQfr/u/QhL8U8DdFhcQ==
-----END CERTIFICATE-----
`

	INVALID_CLIENT_PUB = `-----BEGIN CERTIFICATE-----
MIIBqjCCAU+gAwIBAgIUUz/ze0Cu1D/o2gN//6L6U2a7lkQwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHUm9vdCBDQTAeFw0yNTA0MjIxNTQ2MDNaFw0yNjA0MjIxNTQ2
MDNaMBQxEjAQBgNVBAMMCWZhaWxfdXNlcjBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABHZaJW6qtmg8/Svs7dP0bf3Y2+wkxhK6yRLs6FsJLgHT6OSRzVOiaG+suGUB
zU9/Q925wGbe1pyRkI2dYubMDXCjgYAwfjALBgNVHQ8EBAMCBDAwEwYDVR0lBAww
CgYIKwYBBQUHAwIwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMB0GA1UdDgQW
BBRjv6w1L/zdfldRJb8VMO1WuoFWWDAfBgNVHSMEGDAWgBTC9K+Dk2DoiapITI1W
iQFHhjSJNTAKBggqhkjOPQQDAgNJADBGAiEA6n6iR9sqs+v49E85adenjpFY6OXU
wCn/efr4W3ILissCIQCN/C+XFa5eOCaI1OLqlcLXsonmv9JmuDM14KysLncX4w==
-----END CERTIFICATE-----
`

	INVALID_CLIENT_PRIV = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMUHGbYzzSYNY/p8Vp5Dx37OTAlgbm+RYgI8YAclg+QGoAoGCCqGSM49
AwEHoUQDQgAEdlolbqq2aDz9K+zt0/Rt/djb7CTGErrJEuzoWwkuAdPo5JHNU6Jo
b6y4ZQHNT39D3bnAZt7WnJGQjZ1i5swNcA==
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

	if _, _, err := splitFlagsAndRemoteCommand([]string{"start", "-h"}); err == nil {
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

	if err = start(validCommand); err != nil && gstatus.Code(err) != codes.Unimplemented {
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

	if err := start([]string{"start", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "--", "ls"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Start should behave correctly: %s", err.Error()))
	}

	if err := status([]string{"status", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Status should behave correctly: %s", err.Error()))
	}

	if err := stop([]string{"stop", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Stop should behave correctly: %s", err.Error()))
	}

	if err := tail([]string{"tail", "-s", connAddr, "--ca-file", VALID_CERT_AUTHORITY, "--cert-file", VALID_CLIENT_PUB, "--key-file", VALID_CLIENT_PRIV, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Tail should behave correctly: %s", err.Error()))
	}
}
