package main

import (
	"github.com/richyhbm/teleport-challenge/proto"
	"google.golang.org/grpc"
)

type StreamWriter struct {
	stream grpc.ServerStreamingServer[proto.JobOutputResponse]
}

func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	return len(p), sw.stream.Send(&proto.JobOutputResponse{
		Message: p,
	})
}
