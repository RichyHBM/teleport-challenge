package main

import (
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

//  Datatype to wrap the grpc.ServerStreamingServer in to an io.Writer
type StreamWriter struct {
	stream grpc.ServerStreamingServer[proto.JobOutputResponse]
}

// StreamWriter Write implements the io.Writer, sending any given data
// to the grpc.ServerStreamingServer stream
func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	return len(p), sw.stream.Send(&proto.JobOutputResponse{
		Message: p,
	})
}
