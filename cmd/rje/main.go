/*
rje is a program designed to run commands on an external server by use of a client
both the client and server are provided as a single binary.

When used it will execute a job on a remote system and allow users to query or follow
the output of said job, as well as allowing the user to terminate the job if it does
not terminate naturally.

Usage:

	rje <subcommand> -a <certificate authority file> -k <user/server certificate key> -c <user/server certificate>

Subcommands:

	start --server=<server endpoint> -- <command --arguments>
		Once a job has been started, it will return a job ID to be used with the remaining subcommands

	stop --server=<server endpoint> --job=<job ID>

	status --server=<server endpoint> --job=<job ID>

	tail --server=<server endpoint> --job=<job ID>

	serve --port=4567
*/
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
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

	// Wrapper to print the output of the method, and return any error
	printWrapper := func(funcName func([]string) (string, error)) func([]string) error {
		return func(wrapperArgs []string) error {
			outputString, err := funcName(wrapperArgs)
			if err != nil {
				return err
			}
			fmt.Println(outputString)
			return nil
		}
	}

	// Want start to return just the job id, but print more info
	startWrapper := func(wrapperArgs []string) (string, error) {
		str, err := start(wrapperArgs)
		return fmt.Sprintf("Job created: %s\n", str), err
	}

	// Custom logic for managing subcommands
	commandFunctions := map[string]func([]string) error{
		"start":  printWrapper(startWrapper),
		"stop":   printWrapper(stop),
		"status": printWrapper(status),
		"tail":   tail,
		"serve":  serve,
	}

	// Call subcommand, print out error if any
	subcommand := os.Args[1]
	if function, hasKey := commandFunctions[subcommand]; hasKey {
		if err := function(os.Args[1:]); err != nil {
			// if help arg passed in, dont print error
			if !errors.Is(pflag.ErrHelp, err) {
				fmt.Println(err.Error())
			}
		}
	} else {
		fmt.Println("subcommand not found, run program without arguments to see usage")
	}
}
