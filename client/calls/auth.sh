#!/bin/sh

GRPC_HOST="localhost:50051"
GRPC_METHOD="userpb.UserService/AuthUser"

payload=$(
  cat <<EOF
{
  "token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImxvbDhAa2VrLnJ1IiwiZXhwIjoxNjY1ODMzNjY4LCJpYXQiOjE2NjU4MzI3NjgsImlzcyI6InVzZXJhcHAuc2VydmljZS51c2VyIn0.WKjyn0_hCi3rloeX_S9iWfRHWmGtQZiI-Fw05G4hUh8",
    "valid": true
    }
}
EOF
)

grpcurl -plaintext -emit-defaults \
  -rpc-header 'x-app-name:dev' \
  -d "${payload}" ${GRPC_HOST} ${GRPC_METHOD}
