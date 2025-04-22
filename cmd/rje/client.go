package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/orsinium-labs/cliff"
)

// Start is more complex than other methods as we need to
// isolate program arguments from job command/arguments
func start(args []string) error {
	if len(args) <= 1 {
		return errors.New("program requires more arguments")
	}

	// Figure out where the program arguments end in the args array
	argumentsEndIndex := 1

	if args[1] == "-h" || args[1] == "--help" {
		argumentsEndIndex++
	} else if args[1] == "-s" || args[1] == "--server" {
		if len(args) > 2 {
			argumentsEndIndex += 2
		} else {
			argumentsEndIndex += 1
		}
	}

	// Only pass valid args to cliff, isolate job command + args
	flags, err := cliff.Parse(os.Stderr, args[:argumentsEndIndex], startFlags)
	if err != nil {
		return err
	}

	remoteJob := args[argumentsEndIndex:]
	if len(remoteJob) == 0 {
		return errors.New("no remote job command given")
	}

	fmt.Printf("start not implemented: %s, %s", flags.server, remoteJob)
	return nil
}

func stop(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	fmt.Printf("stop not implemented: %s, %s", flags.server, flags.jobId)
	return nil
}

func status(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	fmt.Printf("status not implemented: %s, %s", flags.server, flags.jobId)
	return nil
}

func tail(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	fmt.Printf("tail not implemented: %s, %s", flags.server, flags.jobId)
	return nil
}
