package main

import (
	"github.com/orsinium-labs/cliff"
)

type StartArgs struct {
	server string
}

type StopStatusTailArgs struct {
	server string
	jobId  string
}

type ServeArgs struct {
	port int
}

func startFlags(args *StartArgs) cliff.Flags {
	return cliff.Flags{
		"server": cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
	}
}

func stopStatusTailFlags(args *StopStatusTailArgs) cliff.Flags {
	return cliff.Flags{
		"server": cliff.F(&args.server, 's', "localhost:4567", "Server address to issue job to"),
		"job-id": cliff.F(&args.jobId, 'j', "", "Job identifier to run command against"),
	}
}

func serveFlags(args *ServeArgs) cliff.Flags {
	return cliff.Flags{
		"port": cliff.F(&args.port, 'p', 4567, "Port for server to listen on"),
	}
}
