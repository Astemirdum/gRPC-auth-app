package handler

import (
	"authapp/authpb"
	"authapp/server"
	"authapp/server/entity"
	"authapp/server/pkg/service"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cacheKey = "lol"
)

type Handler struct {
	s     *service.Service
	cache *server.Cache
	prod  *server.Producer
}

func NewHandler(srv *service.Service, cache *server.Cache, prod *server.Producer) *Handler {
	return &Handler{
		s:     srv,
		cache: cache,
		prod:  prod,
	}
}

func (h *Handler) CreateUser(ctx context.Context, req *authpb.CreateRequest) (*authpb.CreateResponse, error) {
	user := unmarshal(req.GetUser())
	id, err := h.s.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateUser: %v", err)
	}

	// if err := h.prod.Produce(ctx, id); err != nil {
	// 	logrus.Printf("kafka produce: %v", err)
	// 	return &authpb.CreateResponse{Id: int32(id)}, status.Errorf(codes.Internal, "kafka produce: %v", err)
	// }
	if err := h.cache.DeleteCache(ctx, cacheKey); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateUser DeleteCache: %v", err)
	}
	return &authpb.CreateResponse{Id: int32(id)}, nil
}

func (h *Handler) GetAllUser(req *authpb.GetAllRequest, stream authpb.AuthService_GetAllUserServer) error {
	ctx := stream.Context()
	users, err := h.cache.GetCache(ctx, cacheKey)
	if err != nil && err != redis.Nil {
		return status.Errorf(codes.Unavailable, "GetCache: %v", err)
	}
	if err == redis.Nil {
		logrus.Println("Empty cache")
		users, err = h.s.GetAllUser(ctx)
		if err != nil {
			return status.Errorf(codes.Internal, "GetAllUser: %v", err)
		}
		if err := h.cache.SetCache(ctx, cacheKey, users); err != nil {
			return status.Errorf(codes.Internal, "SetCache: %v", err)
		}
	}
	usrs := marshalCollection(users)
	for _, u := range usrs {
		if err := stream.Send(&authpb.GetAllResponse{User: u}); err != nil {
			return status.Errorf(codes.Unavailable, "stream.Send: %v", err)
		}
	}
	return nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *authpb.DeleteRequest) (*authpb.DeleteResponse, error) {
	if _, err := h.s.DeleteUser(ctx, int(req.GetId())); err != nil {
		return &authpb.DeleteResponse{Success: false}, status.Errorf(codes.Internal, "DeleteUser %v", err)
	}
	if err := h.cache.DeleteCache(ctx, cacheKey); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteUser DeleteCache: %v", err)
	}
	return &authpb.DeleteResponse{Success: true}, nil
}

func (h *Handler) AuthUser(ctx context.Context, req *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	token, err := h.s.GenerateToken(ctx, unmarshal(req.GetUser()))
	if err != nil {
		return &authpb.AuthResponse{Token: &authpb.Token{
			Token: "",
			Valid: false,
		}}, status.Errorf(codes.Internal, "AuthUser %v", err)
	}

	return &authpb.AuthResponse{Token: &authpb.Token{
		Token: token,
		Valid: true,
	}}, nil

}

func (h *Handler) ValidateToken(ctx context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	token := req.GetToken().Token
	if _, err := h.s.ParseToken(token); err != nil {
		return nil, status.Errorf(codes.Unavailable, "ValidateToken %v", err)
	}
	return &authpb.ValidateResponse{
		Token: &authpb.Token{
			Token: token,
			Valid: true,
		},
	}, nil
}

func (h *Handler) AuthInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	start := time.Now()
	if info.FullMethod == "/authapp.AuthService/DeleteUser" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "retrieving metadata failed")
		}
		token, ok := md["authorization"]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no auth details supplied")
		}
		email, err := h.s.ParseToken(token[0])

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "token not valid %v", err)
		}
		emailToValid, ok := md["email"]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no email details supplied")
		}
		if email != emailToValid[0] {
			return nil, status.Errorf(codes.InvalidArgument, "email not fit token")
		}
		logrus.Printf("Validate Token email:%s passed: %s", emailToValid[0], token[0])
	}
	reply, err := handler(ctx, req)

	logrus.Printf("request - Method:%s  Duration:%s	Error:%v",
		info.FullMethod,
		time.Since(start),
		err)
	return reply, err
}

func unmarshal(user *authpb.User) *entity.User {
	return &entity.User{
		Id:       int(user.GetId()),
		Email:    user.GetEmail(),
		Password: user.GetPassword(),
	}
}

func marshalCollection(usrs []*entity.User) []*authpb.User {
	users := make([]*authpb.User, len(usrs))
	for _, u := range usrs {
		users = append(users, &authpb.User{
			Id:       int32(u.Id),
			Email:    u.Email,
			Password: u.Password,
		})
	}
	return users
}
