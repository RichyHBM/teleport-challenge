package main

import (
	"fmt"
	"log"
	"testing"
)

func startServer(certFile string, keyFile string, certAuthorityFile string) (func(), error) {
	grpcServer, listener, err := createGrpcServer(5555, certFile, keyFile, certAuthorityFile)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return func() {
		grpcServer.GracefulStop()
		listener.Close()
	}, nil
}

func TestClientArgs(t *testing.T) {
	_, _, err := splitFlagsAndRemoteCommand([]string{"start"})
	if err == nil {
		t.Error("no arguments should fail")
	}

	_, _, err = splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:123"})
	if err == nil {
		t.Error("no remote command separator should fail")
	}

	_, _, err = splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:123", "--"})
	if err == nil {
		t.Error("no remote command should fail")
	}

	_, _, err = splitFlagsAndRemoteCommand([]string{"start", "-h"})
	if err == nil {
		t.Error("help argument should be valid")
	}

	args, remoteJob, err := splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:5555", "--", "ls"})
	if err != nil {
		t.Error(fmt.Sprintf("command should pass: %s", err.Error()))
	} else {
		if len(remoteJob) != 1 {
			t.Error("failed to parse remote job")
		}
		if args.server != "localhost:5555" {
			t.Error("server argument failed to parse")
		}
	}

	args, remoteJob, err = splitFlagsAndRemoteCommand([]string{"start", "-s", "localhost:5555", "--", "ls", "-lsa"})
	if err != nil {
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
	shutdownServer, err := startServer("server.crt", "server.key", "CA.pem")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	err = start([]string{"start", "-s", "localhost:5555", "--cert-file", "client_fail.crt", "--key-file", "client_fail.key", "--", "ls"})
	if err == nil {
		t.Error(fmt.Sprintf("Connection should fail with bad keys"))
	}

	err = start([]string{"start", "-s", "localhost:5555", "--cert-file", "client.crt", "--key-file", "client.key", "--", "ls"})
	if err != nil && err.Error() == "not implemented" {
		t.Error(fmt.Sprintf("Connection should pass with good keys: %s", err.Error()))
	}
}

func TestClientCommands(t *testing.T) {
	shutdownServer, err := startServer("server.crt", "server.key", "CA.pem")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer shutdownServer()

	err = start([]string{"start", "-s", "localhost:5555", "--", "ls"})
	if err != nil && err.Error() == "not implemented" {
		t.Error(fmt.Sprintf("Start should behave correctly: %s", err.Error()))
	}

	err = status([]string{"status", "-s", "localhost:5555", "-j", "123"})
	if err != nil && err.Error() == "not implemented" {
		t.Error(fmt.Sprintf("Status should behave correctly: %s", err.Error()))
	}

	err = stop([]string{"stop", "-s", "localhost:5555", "-j", "123"})
	if err != nil && err.Error() == "not implemented" {
		t.Error(fmt.Sprintf("Stop should behave correctly: %s", err.Error()))
	}

	err = tail([]string{"tail", "-s", "localhost:5555", "-j", "123"})
	if err != nil && err.Error() == "not implemented" {
		t.Error(fmt.Sprintf("Tail should behave correctly: %s", err.Error()))
	}
}

