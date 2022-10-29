#!/bin/sh

GRPC_HOST="localhost:50051"
GRPC_METHOD="userpb.UserService/CreateUser"


d=13
for (( i=0; i < 5; i++ ))
do

payload=$(
  cat <<EOF
{
  "user": {
    "email": "email$d$i",
    "password": "pass"
  }
}
EOF
)

grpcurl -plaintext -emit-defaults \
  -d "${payload}" ${GRPC_HOST} ${GRPC_METHOD}

done