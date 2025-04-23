package main

import (
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

func startServer(certFile string, keyFile string, certAuthorityFile string) (func(), string, error) {
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
	shutdownServer, connAddr, err := startServer("server.crt", "server.key", "CA.pem")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	validCommand := []string{"start", "-s", connAddr, "--cert-file", "client.crt", "--key-file", "client.key", "--", "ls"}
	failCommand := []string{"start", "-s", connAddr, "--cert-file", "client_fail.crt", "--key-file", "client_fail.key", "--", "ls"}

	if err = start(failCommand); err == nil {
		t.Error(fmt.Sprintf("Connection should fail with bad keys"))
	}

	if err = start(validCommand); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Connection should pass with good keys: %s", err.Error()))
	}
}

func TestClientCommands(t *testing.T) {
	shutdownServer, connAddr, err := startServer("server.crt", "server.key", "CA.pem")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	if err := start([]string{"start", "-s", connAddr, "--", "ls"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Start should behave correctly: %s", err.Error()))
	}

	if err := status([]string{"status", "-s", connAddr, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Status should behave correctly: %s", err.Error()))
	}

	if err := stop([]string{"stop", "-s", connAddr, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Stop should behave correctly: %s", err.Error()))
	}

	if err := tail([]string{"tail", "-s", connAddr, "-j", "123"}); err != nil && gstatus.Code(err) != codes.Unimplemented {
		t.Error(fmt.Sprintf("Tail should behave correctly: %s", err.Error()))
	}
}
