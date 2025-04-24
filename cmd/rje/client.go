package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"time"

	"github.com/orsinium-labs/cliff"
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

func splitFlagsAndRemoteCommand(args []string) (StartArgs, []string, error) {
	if len(args) <= 1 {
		return StartArgs{}, nil, errors.New("program requires more arguments")
	}

	// Figure out where the program arguments end in the args array
	argumentsEndIndex := slices.Index(args, "--")

	if argumentsEndIndex < 0 || len(args) < argumentsEndIndex {
		return StartArgs{}, nil, errors.New("incorrect amount of arguments found")
	}

	// Only pass valid args to cliff, isolate job command + args
	flags, err := cliff.Parse(os.Stderr, args[:argumentsEndIndex], startFlags)
	if err != nil {
		return StartArgs{}, nil, err
	}

	remoteJob := args[argumentsEndIndex+1:]
	if len(remoteJob) == 0 {
		return StartArgs{}, nil, errors.New("no remote job command given")
	}
	return flags, remoteJob, nil
}

// Start is more complex than other methods as we need to
// isolate program arguments from job command/arguments
func start(args []string) error {
	flags, remoteJob, err := splitFlagsAndRemoteCommand(args)
	if err != nil {
		return err
	}

	cert, err := loadCerts([]byte(flags.certFile), []byte(flags.keyFile), []byte(flags.caFile))
	if err != nil {
		return err
	}

	grpcClient, err := grpc.NewClient(flags.server, grpc.WithTransportCredentials(cert))
	if err != nil {
		return err
	}

	defer func() {
		if err := grpcClient.Close(); err != nil {
			log.Printf("Unable to close gRPC channel %v\n", err)
		}
	}()

	// Create the gRPC client
	jobServiceClient := proto.NewJobsServiceClient(grpcClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create account
	jobResponse, err := jobServiceClient.Start(ctx, &proto.JobStartRequest{Command: remoteJob})
	if err != nil {
		return err
	}

	fmt.Printf("Job created: %s\n", jobResponse.JobId)
	return nil
}

func stop(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	cert, err := loadCerts([]byte(flags.certFile), []byte(flags.keyFile), []byte(flags.caFile))
	if err != nil {
		return err
	}

	grpcClient, err := grpc.NewClient(flags.server, grpc.WithTransportCredentials(cert))
	if err != nil {
		return err
	}

	defer func() {
		if err = grpcClient.Close(); err != nil {
			log.Printf("Unable to close gRPC channel %v", err)
		}
	}()

	// Create the gRPC client
	jobServiceClient := proto.NewJobsServiceClient(grpcClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create account
	jobResponse, err := jobServiceClient.Stop(ctx, &proto.JobIdRequest{JobId: flags.jobId})
	if err != nil {
		return err
	}

	fmt.Printf("Job ended with: %d, was force ended: %t\n", jobResponse.ExitCode, jobResponse.ForceEnded)
	return nil
}

func status(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	cert, err := loadCerts([]byte(flags.certFile), []byte(flags.keyFile), []byte(flags.caFile))
	if err != nil {
		return err
	}

	grpcClient, err := grpc.NewClient(flags.server, grpc.WithTransportCredentials(cert))
	if err != nil {
		return err
	}

	defer func() {
		if err = grpcClient.Close(); err != nil {
			log.Printf("Unable to close gRPC channel %v", err)
		}
	}()

	// Create the gRPC client
	jobServiceClient := proto.NewJobsServiceClient(grpcClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create account
	jobResponse, err := jobServiceClient.Status(ctx, &proto.JobIdRequest{JobId: flags.jobId})
	if err != nil {
		return err
	}

	switch jobResponse.JobStatus {
	case proto.JobStatus_JobStatus_RUNNING:
		fmt.Println("Job running")
	case proto.JobStatus_JobStatus_ENDED:
		fmt.Printf("Job ended with code with code: %d\n", jobResponse.ExitCode)
	case proto.JobStatus_JobStatus_FORCE_ENDED:
		fmt.Printf("Job force ended with code: %d\n", jobResponse.ExitCode)
	}

	return nil
}

func tail(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, stopStatusTailFlags)
	if err != nil {
		return err
	}

	cert, err := loadCerts([]byte(flags.certFile), []byte(flags.keyFile), []byte(flags.caFile))
	if err != nil {
		return err
	}

	grpcClient, err := grpc.NewClient(flags.server, grpc.WithTransportCredentials(cert))
	if err != nil {
		return err
	}

	defer func() {
		if err = grpcClient.Close(); err != nil {
			log.Printf("Unable to close gRPC channel %v", err)
		}
	}()

	// Create the gRPC client
	jobServiceClient := proto.NewJobsServiceClient(grpcClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create account
	jobResponse, err := jobServiceClient.Tail(ctx, &proto.JobIdRequest{JobId: flags.jobId})
	if err != nil {
		return err
	}

	for {
		if jobTailResponse, err := jobResponse.Recv(); err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		} else {
			fmt.Println(jobTailResponse)
		}
	}
}
