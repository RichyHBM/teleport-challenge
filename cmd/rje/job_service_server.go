package main

import (
	"context"
	"errors"

	"github.com/richyhbm/teleport-challenge/pkg/rje"
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

// Custom GRPC Server implementation
type jobServiceServer struct {
	proto.UnimplementedJobsServiceServer
	remoteJobRunner rje.RemoteJobRunner
}

// Start checks the passed in request data and calls the library Start method to initiate a new job
func (jSS *jobServiceServer) Start(ctx context.Context, req *proto.JobStartRequest) (*proto.JobStartResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	jobId, isRunning, err := jSS.remoteJobRunner.Start(req.Command)
	if err != nil {
		return nil, err
	}

	status := proto.JobStartStatus_JobStartStatus_RUNNING
	if !isRunning {
		status = proto.JobStartStatus_JobStartStatus_EXITED_INSTANTLY
	}

	return &proto.JobStartResponse{
		JobId:  jobId,
		Status: status,
	}, nil
}

// Stop checks the passed in request data and calls the library Stop method to terminate a running job
func (jSS *jobServiceServer) Stop(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStopResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	exitCode, exitedClean, err := jSS.remoteJobRunner.Stop(req.JobId)
	if err != nil {
		return nil, err
	}

	return &proto.JobStopResponse{
		ExitCode:   int32(exitCode),
		ForceEnded: !exitedClean,
	}, nil
}

// Status checks the passed in request data and calls the library Status method to query a job
func (jSS *jobServiceServer) Status(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStatusResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	isRunning, processState, err := jSS.remoteJobRunner.Status(req.JobId)
	if err != nil {
		return nil, err
	}

	var status proto.JobStatus
	exitCode := -1

	if isRunning {
		status = proto.JobStatus_JobStatus_RUNNING
	} else {
		status = proto.JobStatus_JobStatus_ENDED
	}

	if processState != nil {
		exitCode = processState.ExitCode()
		if processState.Exited() {
			status = proto.JobStatus_JobStatus_ENDED
		} else {
			status = proto.JobStatus_JobStatus_FORCE_ENDED
		}
	}

	return &proto.JobStatusResponse{
		ExitCode:  int32(exitCode),
		JobStatus: status,
	}, nil
}

// Tail checks the passed in request data and calls the library Tail, streaming back any output, and blocking until the job has ended
func (jSS *jobServiceServer) Tail(req *proto.JobIdRequest, stream grpc.ServerStreamingServer[proto.JobOutputResponse]) error {
	if req == nil {
		return errors.New("empty request")
	}

	return jSS.remoteJobRunner.Tail(req.JobId, &StreamWriter{stream: stream})
}
