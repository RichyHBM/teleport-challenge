package main

import (
	"context"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	grpc_status "google.golang.org/grpc/status"
)

var (
	ErrUnAuth                     = grpc_status.Error(codes.PermissionDenied, "unauthorized")
	authData  map[string][]string = map[string][]string{
		"root":         {"*"},
		"valid_client": {"echo", "cat", "ls", "tail", "sleep"},
	}
)

// This would load auth from somewhere
func getAuthData() map[string][]string {
	return authData
}

func IsAuthorized(ctx context.Context, command string) error {
	peer, ok := peer.FromContext(ctx)
	if !ok || peer.AuthInfo == nil {
		return ErrUnAuth
	}

	var user string

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	if tlsInfo.State.VerifiedChains != nil && len(tlsInfo.State.VerifiedChains) > 0 && len(tlsInfo.State.VerifiedChains[0]) > 0 {
		user = tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	}

	authData := getAuthData()
	authorizations, isKnown := authData[user]
	if !isKnown {
		return ErrUnAuth
	}

	// Special case for root
	if len(authorizations) == 1 && authorizations[0] == "*" {
		return nil
	}

	if slices.Index(authorizations, command) >= 0 {
		return nil
	}

	return ErrUnAuth
}
