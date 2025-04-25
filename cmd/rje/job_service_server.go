package main

import (
	"context"
	"errors"

	"github.com/richyhbm/teleport-challenge/pkg/rje"
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

type jobServiceServer struct {
	proto.UnimplementedJobsServiceServer
	remoteJobRunner rje.RemoteJobRunner
}

func (jSS *jobServiceServer) Start(ctx context.Context, req *proto.JobStartRequest) (*proto.JobStartResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	jobId, foundExec, isRunning, err := jSS.remoteJobRunner.Start(req.Command)
	if err != nil {
		return nil, err
	}

	status := proto.JobStartStatus_JobStartStatus_RUNNING
	if !isRunning {
		status = proto.JobStartStatus_JobStartStatus_EXITED_INSTANTLY
	}

	if !foundExec {
		status = proto.JobStartStatus_JobStartStatus_COMMAND_NOT_FOUND
	}

	return &proto.JobStartResponse{
		JobId:  jobId,
		Status: status,
	}, nil
}

func (jSS *jobServiceServer) Stop(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStopResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	exitCode, forceKill, err := jSS.remoteJobRunner.Stop(req.JobId)
	if err != nil {
		return nil, err
	}

	return &proto.JobStopResponse{
		ExitCode:   int32(exitCode),
		ForceEnded: forceKill,
	}, nil
}

func (jSS *jobServiceServer) Status(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStatusResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	processState, err := jSS.remoteJobRunner.Status(req.JobId)
	if err != nil {
		return nil, err
	}

	status := proto.JobStatus_JobStatus_RUNNING
	exitCode := -1

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

func (jSS *jobServiceServer) Tail(req *proto.JobIdRequest, stream grpc.ServerStreamingServer[proto.JobOutputResponse]) error {
	if req == nil {
		return errors.New("empty request")
	}

	return jSS.UnimplementedJobsServiceServer.Tail(req, stream)
}
