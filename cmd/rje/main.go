package main

import (
	"fmt"
	"os"
)

var helpMessage = `rje - Remote Job Executor
Program used to issue jobs, external commands, remotely from client:
Subcommands:
	serve 	act as a server

	start  	issue a start command to the server
	stop 	issue a stop command to stop a remote job
	status 	request status of executed job
	tail 	follow output of job until termination

All subcommands accept -h/--help for further information on functionality and arguments
	rje <subcommand> -h
`

func main() {
	// Custom help message for overall program usage, ensure command called with more than 0 args
	// arg[0] is program path
	if len(os.Args) <= 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(helpMessage)
		return
	}

	// Custom logic for managing subcommands
	commandFunctions := map[string]func([]string) error{
		"start":  start,
		"stop":   stop,
		"status": status,
		"tail":   tail,
		"serve":  serve,
	}

	// Call subcommand, print out error if any
	subcommand := os.Args[1]
	if function, hasKey := commandFunctions[subcommand]; hasKey {
		if err := function(os.Args[1:]); err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("subcommand not found, run program without arguments to see usage")
	}
}
