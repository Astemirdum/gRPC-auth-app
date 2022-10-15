package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

func TimingInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	fmt.Printf(`--
	call=%v
	req=%#v
	reply=%#v
	time=%v
	err=%v
`, method, req, reply, time.Since(start), err)
	return err
}

type tokenAuthCreds struct {
	token string
}

func (t *tokenAuthCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": t.token,
	}, nil
}

func (t *tokenAuthCreds) RequireTransportSecurity() bool {
	return false
}

func NewTokenAuthCreds(token string) *tokenAuthCreds {
	return &tokenAuthCreds{token}
}
