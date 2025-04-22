package main

import (
	"fmt"
	"testing"
)

func TestClientArgs(t *testing.T) {
	err := start([]string{"start"})
	if err == nil {
		t.Error("no arguments should fail")
	}

	err = start([]string{"start", "-s", "localhost:123"})
	if err == nil {
		t.Error("no remote command should fail")
	}

	err = start([]string{"start", "-h"})
	if err == nil {
		t.Error("help argument should be valid")
	}

	err = start([]string{"start", "ls"})
	if err != nil {
		t.Error(fmt.Sprintf("command should pass: %s", err.Error()))
	}

	err = start([]string{"start", "ls", "-lsa"})
	if err != nil {
		t.Error(fmt.Sprintf("multiple arguments should pass%s", err.Error()))
	}

	err = start([]string{"start", "-s", "localhost:123", "ls", "-lsa"})
	if err != nil {
		t.Error(fmt.Sprintf("full usage should pass%s", err.Error()))
	}
}
