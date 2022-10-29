package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"google.golang.org/grpc/reflection"

	"github.com/Astemirdum/user-app/server/internal/broker"
	"github.com/Astemirdum/user-app/server/internal/cache"
	"github.com/Astemirdum/user-app/server/internal/config"
	"github.com/Astemirdum/user-app/server/internal/handler"
	"github.com/Astemirdum/user-app/server/internal/repository"
	"github.com/Astemirdum/user-app/server/internal/service"
	"github.com/Astemirdum/user-app/server/internal/storage"
	"github.com/Astemirdum/user-app/userpb"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	configPath := pflag.StringP("config", "c", "config.yml", "config path")
	pflag.Parse()

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("load envs from .env  %v", err)
	}

	cfg := config.ReadConfigYML(*configPath)

	db, err := repository.NewPostgresDB(&cfg.Database)
	if err != nil {
		logrus.Fatalf("db init: %v", err)
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second)
	defer cancelFn()
	red, err := cache.NewCache(ctx,
		net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		cfg.Redis.Password)
	if err != nil {
		logrus.Fatalf("redis init: %v", err)
	}

	br, err := broker.NewBroker(&cfg.Kafka)
	if err != nil {
		logrus.Fatal(err)
	}
	producer, err := broker.NewProducer(&cfg.Kafka)
	if err != nil {
		logrus.Fatal(err)
	}

	st, err := storage.NewStorage(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	srv := handler.NewHandler(services, red, producer)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(srv.AuthInterceptor),
		grpc.StreamInterceptor(handler.LogInterceptor))
	// grpc.Creds(credentials.NewTLS(&tls.Config{}))

	userpb.RegisterUserServiceServer(s, srv)

	reflection.Register(s)

	grpcAddr := net.JoinHostPort(cfg.Grpc.Host, cfg.Grpc.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logrus.Fatalf("unable to listen on %s: %v",
			grpcAddr, err)
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("failed gRPC Server: %v", err)
		}
	}()
	logrus.Infof("Server successfully has been started on %s", grpcAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	logrus.Println("Graceful shutdown")

	_ = db.Close()
	_ = red.Client.Close()
	_ = lis.Close()
	_ = st.Close()
	_ = producer.Close()
	_ = br.Close()
	logrus.Println("END GAME...")
}
