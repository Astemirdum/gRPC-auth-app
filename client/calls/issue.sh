#!/bin/sh

GRPC_HOST="localhost:50051"
GRPC_METHOD="userpb.UserService/IssueToken"

payload=$(
  cat <<EOF
{
  "user": {
    "email": "lol8@kek.ru",
    "password": "lol8"
  }
}
EOF
)

grpcurl -plaintext -emit-defaults \
  -d "${payload}" ${GRPC_HOST} ${GRPC_METHOD}
