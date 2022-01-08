package main

import (
	"authapp/authpb"
	"context"
	"fmt"
	"time"

	"authapp/client/service"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func timingInterceptor(
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

func NewtokenAuthCreds(token string) *tokenAuthCreds {
	return &tokenAuthCreds{token}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := initConfig(); err != nil {
		logrus.Fatalf("initConfigs %s", err.Error())
	}
	// call cs.AuthUser(ctx, user) to get token for deletion authority
	token := "userToken"

	grpcAuth := NewtokenAuthCreds(token)

	cc, err := grpc.Dial(viper.GetString("auth-service.addr"),
		grpc.WithInsecure(), /// grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(grpcAuth),
		grpc.WithUnaryInterceptor(timingInterceptor),
	)
	if err != nil {
		logrus.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	sc := authpb.NewAuthServiceClient(cc)

	cs := service.NewClientService(sc)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.Pairs(
		"api_req_id", "1o1",
		"subsystem", "default_client",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	user := &authpb.User{
		Email:    "lol13@kek.ru",
		Password: "lol13",
	}
	// create user
	id, err := cs.CreateUser(ctx, user)
	if err != nil {
		logrus.Println(err)
	}
	fmt.Println(id)

	// // delete user
	md = metadata.Pairs("email", "lol10@kek.ru")
	ctx = metadata.NewOutgoingContext(ctx, md)
	if _, err := cs.DeleteUser(ctx, 13); err != nil {
		logrus.Println(err)
	} else {
		fmt.Println("user deleted")
	}

	// Auth User
	token, err = cs.AuthUser(ctx, user)
	if err != nil {
		logrus.Println(err)
	}
	fmt.Println("token:", token)

	// ////list users
	users, err := cs.GetAllUser(ctx)
	if err != nil {
		logrus.Println(err)
	}
	fmt.Println(users)

	//// Validate Token
	if _, err := cs.ValidateToken(ctx, token); err != nil {
		logrus.Println(err)
	}
	fmt.Println("token valid: ok")

}

func initConfig() error {
	viper.AddConfigPath("../server/configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
