package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/orsinium-labs/cliff"
	"github.com/richyhbm/teleport-challenge/pkg/rje"
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

// Instanciates a new GRPC server, registering the RemoteJobService server to fulfill the
// server actions
func createGrpcServer(port int, certFile []byte, keyFile []byte, certAuthorityFile []byte, useCgroups bool) (*grpc.Server, net.Listener, *jobServiceServer, error) {
	// Create a listener that listens to localhost
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, nil, nil, err
	}

	tlsCredentials, err := loadCerts(certFile, keyFile, certAuthorityFile)
	if err != nil {
		return nil, nil, nil, err
	}

	grpcServer := grpc.NewServer(grpc.Creds(tlsCredentials))
	jobService := &jobServiceServer{useCgroups: useCgroups}
	proto.RegisterJobsServiceServer(grpcServer, jobService)

	return grpcServer, listener, jobService, nil
}

// Creates a new server instance and runs it, listening on the port provided in the arguments
// This method will block indefinitely
func serve(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, serveFlags)
	if err != nil {
		return err
	}

	// TODO: Dont want to mess with user machine for challenge, so just error out if not supported
	// In prod maybe this could edit the files, but for this only modify for sub-cgroups
	if !flags.skipCgroups {
		if err := rje.CheckCgroupSupportsEntries(); err != nil {
			return err
		}
	}

	grpcServer, listener, jobService, err := createGrpcServer(flags.port, []byte(flags.certFile), []byte(flags.keyFile), []byte(flags.caFile), !flags.skipCgroups)
	if err != nil {
		return err
	}
	defer grpcServer.Stop()

	var parentCGroup *rje.CGroup

	if !flags.skipCgroups {
		parentCGroup, err = rje.SetupCGroupFromName("remote-job-challenge", false)
		if err != nil {
			return err
		}
	}

	// Listen for sig interrupts and properly cleanup
	go func() {
		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt)
		<-interruptChan
		cleanup(grpcServer, parentCGroup)
	}()

	defer func() {
		timer := time.AfterFunc(time.Second*10, func() {
			fmt.Println("Server couldn't stop gracefully in time. Doing force stop.")
			grpcServer.Stop()
		})
		defer timer.Stop()
		grpcServer.GracefulStop()
		jobService.Close()
		if parentCGroup != nil {
			parentCGroup.Close()
		}
	}()

	log.Printf("Starting up on %s", listener.Addr().String())
	return grpcServer.Serve(listener)
}

func cleanup(grpcServer *grpc.Server, parentCGroup *rje.CGroup) {
	timer := time.AfterFunc(time.Second, func() {
		fmt.Println("Server couldn't stop gracefully in time. Doing force stop.")
		grpcServer.Stop()
	})
	defer timer.Stop()
	grpcServer.GracefulStop()
}
