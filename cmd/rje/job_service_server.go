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

	if len(req.Command) == 0 {
		return nil, errors.New("no command sent")
	}

	if err := IsAuthorized(ctx, req.Command[0]); err != nil {
		return nil, ErrUnAuth
	}

	return jSS.UnimplementedJobsServiceServer.Start(ctx, req)
}

func (jSS *jobServiceServer) Stop(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStopResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	return jSS.UnimplementedJobsServiceServer.Stop(ctx, req)
}

func (jSS *jobServiceServer) Status(ctx context.Context, req *proto.JobIdRequest) (*proto.JobStatusResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	return jSS.UnimplementedJobsServiceServer.Status(ctx, req)
}

func (jSS *jobServiceServer) Tail(req *proto.JobIdRequest, stream grpc.ServerStreamingServer[proto.JobOutputResponse]) error {
	if req == nil {
		return errors.New("empty request")
	}

	return jSS.UnimplementedJobsServiceServer.Tail(req, stream)
}
