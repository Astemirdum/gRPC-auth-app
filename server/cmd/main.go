package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"authapp/authpb"
	"authapp/server"
	"authapp/server/pkg/handler"
	"authapp/server/pkg/repository"
	"authapp/server/pkg/service"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := initConfig(); err != nil {
		logrus.Fatalf("initConfigs %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("load envs from .env  %s", err.Error())
	}

	db, err := server.NewPostgresDB(&server.ConfigDB{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		User:     viper.GetString("db.user"),
		Password: os.Getenv("DB_PASSWORD"),
		DBname:   viper.GetString("db.dbname"),
	})
	if err != nil {
		logrus.Fatalf("db init: %s", err.Error())
	}
	db.Exec(`create table if not exists users (
		id serial primary key,
		email varchar(50) NOT NULL UNIQUE,
		password_hash varchar(100) NOT NULL 
	);`)
	cache, err := server.NewRedisClient(viper.GetString("redis.addr"), viper.GetString("redis.passwd"))
	if err != nil {
		logrus.Fatalf("redis init: %s", err.Error())
	}

	producer := server.NewKafkaProducer(viper.GetString("kafka.topic"), viper.GetString("kafka.addr"))

	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	srv := handler.NewHandler(services, cache, producer)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(srv.AuthInterceptor),
		grpc.StreamInterceptor(streamInterceptor))
	// grpc.Creds(credentials.NewTLS(&tls.Config{}))
	authpb.RegisterAuthServiceServer(s, srv)

	lis, err := net.Listen("tcp", viper.GetString("auth-service.addr"))
	if err != nil {
		logrus.Fatalf("unable to listen on port :50051: %v", err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("failed gRPC Server: %v", err)
		}
	}()

	logrus.Println("Server succesfully has been started on port :50051")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	logrus.Println("Graceful shutdown")

	db.Close()
	cache.Client.Close()
	producer.Prod.Close()
	lis.Close()
	logrus.Println("END GAME...")
}

func initConfig() error {
	viper.AddConfigPath("../configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func streamInterceptor(srv interface{},
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
