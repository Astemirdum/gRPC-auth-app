package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Astemirdum/user-app/server"
	"github.com/Astemirdum/user-app/server/pkg/handler"
	"github.com/Astemirdum/user-app/server/pkg/repository"
	"github.com/Astemirdum/user-app/server/pkg/service"
	"github.com/Astemirdum/user-app/userpb"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := initConfig(); err != nil {
		logrus.Fatalf("initConfigs %v", err)
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("load envs from .env  %v", err)
	}

	db, err := server.NewPostgresDB(&server.ConfigDB{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetInt("db.port"),
		Username: viper.GetString("db.user"),
		Password: os.Getenv("DB_PASSWORD"),
		NameDB:   viper.GetString("db.dbname"),
	})
	if err != nil {
		logrus.Fatalf("db init: %v", err)
	}

	cache, err := server.NewRedisClient(viper.GetString("redis.addr"), viper.GetString("redis.passwd"))
	if err != nil {
		logrus.Fatalf("redis init: %v", err)
	}

	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	srv := handler.NewHandler(services, cache)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(srv.AuthInterceptor),
		grpc.StreamInterceptor(logInterceptor))
	//grpc.Creds(credentials.NewTLS(&tls.Config{}))
	userpb.RegisterUserServiceServer(s, srv)

	lis, err := net.Listen("tcp", viper.GetString("user-service.addr"))
	if err != nil {
		logrus.Fatalf("unable to listen on %s: %v", viper.GetString("user-service.addr"), err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("failed gRPC Server: %v", err)
		}
	}()
	logrus.Infof("Server successfully has been started on %s", viper.GetString("user-service.addr"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	logrus.Println("Graceful shutdown")

	_ = db.Close()
	_ = cache.Client.Close()
	_ = lis.Close()
	logrus.Println("END GAME...")
}

func initConfig() error {
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func logInterceptor(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {

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
