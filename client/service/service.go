package service

import (
	"context"
	"fmt"
	"io"

	"github.com/Astemirdum/user-app/authpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClientService struct {
	cs authpb.AuthServiceClient
}

func NewClientService(cs authpb.AuthServiceClient) *ClientService {
	return &ClientService{cs}
}

func (c *ClientService) CreateUser(ctx context.Context, user *authpb.User) (int, error) {
	res, err := c.cs.CreateUser(ctx, &authpb.CreateRequest{User: user})
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				return 0, fmt.Errorf("deadline exceeded")
			}
			return 0, fmt.Errorf("unexpected error: %s", statusErr.Message())
		}
		return 0, fmt.Errorf("error while calling: %s", err.Error())
	}
	return int(res.Id), nil
}

func (c *ClientService) GetAllUser(ctx context.Context) ([]*authpb.User, error) {

	stream, err := c.cs.GetAllUser(ctx, &authpb.GetAllRequest{})
	if err != nil {
		return nil, fmt.Errorf("reading stream GetAllUser %v", err)
	}
	users := make([]*authpb.User, 0)
	for {
		mesg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading stream %v", err)
		}
		users = append(users, mesg.GetUser())
	}
	return users, nil
}

func (c *ClientService) DeleteUser(ctx context.Context, id int) (bool, error) {
	res, err := c.cs.DeleteUser(ctx, &authpb.DeleteRequest{Id: int32(id)})
	if err != nil {
		return false, err
	}
	return res.Success, nil
}

func (c *ClientService) AuthUser(ctx context.Context, user *authpb.User) (string, error) {
	res, err := c.cs.AuthUser(ctx, &authpb.AuthRequest{User: user})
	if err != nil {
		return "", fmt.Errorf("error AuthUser %v", err)
	}
	token := res.Token
	return token.Token, nil
}

func (c *ClientService) ValidateToken(ctx context.Context, token string) (bool, error) {
	res, err := c.cs.ValidateToken(ctx, &authpb.ValidateRequest{
		Token: &authpb.Token{Token: token},
	})
	if err != nil {
		return false, fmt.Errorf("error ValidateToken %v", err)
	}
	if !res.GetToken().Valid {
		return false, fmt.Errorf("not valid token")
	}

	return true, nil
}
