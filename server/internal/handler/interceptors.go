package handler

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const deleteMethod = "/userpb.UserService/DeleteUser"

func (h *Handler) AuthInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	if info.FullMethod == deleteMethod {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "retrieving metadata failed")
		}
		authMD, ok := md["authorization"]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no auth details supplied")
		}
		headerToken := authMD[0]
		if headerToken == "" {
			return nil, status.Errorf(codes.Unauthenticated, "empty authMD")
		}

		headerTokenParts := strings.Split(headerToken, " ")
		logrus.Debugf("%v", headerTokenParts)
		if len(headerTokenParts) != 2 || headerTokenParts[0] != "Bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		token := headerTokenParts[1]
		if len(token) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "empty token")
		}
		email, err := h.s.ParseToken(token)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "token not valid %v", err)
		}
		logrus.Debugf("email auth - email:%s", email)
	}
	reply, err := handler(ctx, req)

	logrus.Debugf("request - Method:%s  Duration:%s	Error:%v",
		info.FullMethod,
		time.Since(start),
		err)
	return reply, err
}

func LogInterceptor(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ss.Context())

	err := handler(srv, ss)

	logrus.Printf("request - Method:%s Duration:%s MD:%v Error:%v ",
		info.FullMethod,
		time.Since(start),
		md,
		err)
	return err
}
