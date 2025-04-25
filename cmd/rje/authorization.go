package main

import (
	"slices"

	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
)

var (
	ErrUnAuth                     = grpc_status.Error(codes.PermissionDenied, "unauthorized")
	authData  map[string][]string = map[string][]string{
		"root":         {"*"},
		"valid_client": {"echo", "cat", "ls"},
	}
)

// This would load auth from somewhere
func getAuthData() map[string][]string {
	return authData
}

func HasAnyAuthorization(user string) bool {
	authData := getAuthData()
	_, isKnown := authData[user]
	return isKnown
}

func IsAuthorized(user string, command string) bool {
	authData := getAuthData()
	authorizations, isKnown := authData[user]
	if !isKnown {
		return false
	}

	// Special case for root
	if len(authorizations) == 1 && authorizations[0] == "*" {
		return true
	}

	return slices.Index(authorizations, command) >= 0
}
