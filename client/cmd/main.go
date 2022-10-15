package main

import (
	"context"
	"github.com/Astemirdum/user-app/client"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	"github.com/Astemirdum/user-app/client/service"
	"github.com/Astemirdum/user-app/userpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// call cs.AuthUser(ctx, user) to get token for deletion authority

	//TODO: issue token for user delete
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImxvbDhAa2VrLnJ1IiwiZXhwIjoxNjY1ODMzNjY4LCJpYXQiOjE2NjU4MzI3NjgsImlzcyI6InVzZXJhcHAuc2VydmljZS51c2VyIn0.WKjyn0_hCi3rloeX_S9iWfRHWmGtQZiI-Fw05G4hUh8"
	grpcAuth := service.NewTokenAuthCreds(token)

	cfg := client.ReadConfigYML("config.yml")
	cc, err := grpc.Dial(cfg.Service.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(grpcAuth),
		grpc.WithUnaryInterceptor(service.TimingInterceptor),
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
		Email:    "lol8@kek.ru",
		Password: "lol8",
	}
	// create user
	id, err := cs.CreateUser(ctx, user)
	if err != nil {
		logrus.Fatalf("createUser %v", err)
	}
	logrus.Println(id)

	// list users
	users, err := cs.GetAllUser(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Println(users)

	// IssueToken
	user = &userpb.User{
		Email:    "lol8@kek.ru",
		Password: "lol8",
	}
	token, err = cs.IssueToken(ctx, user)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("token:", token)

	// delete user
	// to del user -> IssueToken for user -> add to grpc.WithPerRPCCredentials(grpcAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)
	if err := cs.DeleteUser(ctx, 10); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("user deleted")

	//// Validate Token
	if err := cs.AuthUser(ctx, token); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("token valid: ok")

}
