#!/bin/sh

GRPC_HOST="localhost:50051"
GRPC_METHOD="userpb.UserService/GetAllUser"

grpcurl -plaintext -emit-defaults \
   ${GRPC_HOST} ${GRPC_METHOD}
