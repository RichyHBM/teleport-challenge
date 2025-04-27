package main

import (
	"github.com/orsinium-labs/cliff"
)

type StartArgs struct {
	server   string
	caFile   string
	keyFile  string
	certFile string
}

type StopStatusTailArgs struct {
	server   string
	jobId    string
	caFile   string
	keyFile  string
	certFile string
}

type ServeArgs struct {
	port        int
	caFile      string
	keyFile     string
	certFile    string
	skipCgroups bool
}

func startFlags(args *StartArgs) cliff.Flags {
	return cliff.Flags{
		"server":    cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
		"ca-file":   cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":  cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file": cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
	}
}

func stopStatusTailFlags(args *StopStatusTailArgs) cliff.Flags {
	return cliff.Flags{
		"server":    cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
		"job-id":    cliff.F(&args.jobId, 'j', "", "Job identifier to run command against"),
		"ca-file":   cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":  cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file": cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
	}
}

func serveFlags(args *ServeArgs) cliff.Flags {
	return cliff.Flags{
		"port":         cliff.F(&args.port, 'p', 4567, "Port for server to listen on"),
		"ca-file":      cliff.F(&args.caFile, 'a', "", "Certificate Authority file contents"),
		"key-file":     cliff.F(&args.keyFile, 'k', "", "Certificate Key file contents"),
		"cert-file":    cliff.F(&args.certFile, 'c', "", "Certificate file contents"),
		"skip-cgroups": cliff.F(&args.skipCgroups, 's', false, "Skip creating of cgroups"),
	}
}
