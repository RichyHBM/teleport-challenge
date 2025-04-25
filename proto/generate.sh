#!/bin/sh

protoc --proto_path=./ \
  --go_out=paths=source_relative:./ \
  --go-grpc_out=paths=source_relative,require_unimplemented_servers=false:./ \
  ./*.proto