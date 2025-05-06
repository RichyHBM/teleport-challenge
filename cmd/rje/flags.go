package main

import (
	"github.com/orsinium-labs/cliff"
)

// Arguments the user can pass in when using the start subcommand
type StartArgs struct {
	server   string
	caFile   string
	keyFile  string
	certFile string
}

// Arguments the user can pass in when using the stop, status, or tail subcommand
type StopStatusTailArgs struct {
	server   string
	jobId    string
	caFile   string
	keyFile  string
	certFile string
}

// Arguments the user can pass in when using the serve subcommand
type ServeArgs struct {
	port     int
	caFile   string
	keyFile  string
	certFile string
}

// Function to parse args in to StartArgs
func startFlags(args *StartArgs) cliff.Flags {
	return cliff.Flags{
		"server":    cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
		"ca-file":   cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":  cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file": cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
	}
}

// Function to parse args in to StopStatusTailArgs
func stopStatusTailFlags(args *StopStatusTailArgs) cliff.Flags {
	return cliff.Flags{
		"server":    cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
		"job-id":    cliff.F(&args.jobId, 'j', "", "Job identifier to run command against"),
		"ca-file":   cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":  cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file": cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
	}
}

// Function to parse args in to ServeArgs
func serveFlags(args *ServeArgs) cliff.Flags {
	return cliff.Flags{
		"port":      cliff.F(&args.port, 'p', 4567, "Port for server to listen on"),
		"ca-file":   cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":  cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file": cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
	}
}
