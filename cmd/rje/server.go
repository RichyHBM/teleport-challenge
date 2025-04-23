package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/orsinium-labs/cliff"
	"github.com/richyhbm/teleport-challenge/certs"
	"github.com/richyhbm/teleport-challenge/pkg/rje"
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

func createGrpcServer(port int, certFile string, keyFile string, certAuthorityFile string) (*grpc.Server, net.Listener, error) {
	// Create a listener that listens to localhost
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, nil, err
	}

	tlsCredentials, err := certs.LoadCerts(certFile, keyFile, certAuthorityFile, false)
	if err != nil {
		return nil, nil, err
	}

	grpcServer := grpc.NewServer(grpc.Creds(tlsCredentials))
	proto.RegisterJobsServiceServer(grpcServer, &jobServiceServer{})

	return grpcServer, listener, nil
}

func serve(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, serveFlags)
	if err != nil {
		return err
	}

	grpcServer, listener, err := createGrpcServer(flags.port, flags.certFile, flags.keyFile, flags.caFile)
	if err != nil {
		return err
	}

	defer func() {
		grpcServer.GracefulStop()
		listener.Close()
	}()

	log.Printf("Starting up on %s", listener.Addr().String())
	return grpcServer.Serve(listener)
}

type jobServiceServer struct {
	proto.UnimplementedJobsServiceServer
	remoteJobRunner rje.RemoteJobRunner
}

func (jSS *jobServiceServer) Start(ctx context.Context, req *proto.JobStartRequest) (*proto.JobStartResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
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
