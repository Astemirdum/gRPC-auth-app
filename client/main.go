package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Astemirdum/user-app/client/service"
	"github.com/Astemirdum/user-app/userpb"
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

func NewTokenAuthCreds(token string) *tokenAuthCreds {
	return &tokenAuthCreds{token}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := initConfig(); err != nil {
		logrus.Fatalf("initConfigs %s", err.Error())
	}
	// call cs.AuthUser(ctx, user) to get token for deletion authority
	token := "userToken"

	grpcAuth := NewTokenAuthCreds(token)

	cc, err := grpc.Dial(viper.GetString("user-service.addr"),
		grpc.WithInsecure(), /// grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(grpcAuth),
		grpc.WithUnaryInterceptor(timingInterceptor),
	)
	if err != nil {
		logrus.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	cs := service.NewClientService(userpb.NewUserServiceClient(cc))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.Pairs(
		"api_req_id", "1o1",
		"subsystem", "default_client",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	user := &userpb.User{
		Email:    "lol6@kek.ru",
		Password: "lol6",
	}
	// create user
	id, err := cs.CreateUser(ctx, user)
	if err != nil {
		logrus.Fatalf("createUser %v", err)
	}
	logrus.Println(id)

	// delete user
	md = metadata.Pairs("email", "lol12@kek.ru") //  md for authorize del option
	ctx = metadata.NewOutgoingContext(ctx, md)
	if err := cs.DeleteUser(ctx, 2); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("user deleted")

	// list users
	users, err := cs.GetAllUser(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Println(users)

	// IssueToken
	token, err = cs.IssueToken(ctx, user)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("token:", token)

	//// Validate Token
	if err := cs.AuthUser(ctx, token); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("token valid: ok")
}

func initConfig() error {
	viper.AddConfigPath("../server/configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
