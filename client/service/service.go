package service

import (
	"context"
	"fmt"
	"io"

	"github.com/Astemirdum/user-app/userpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClientService struct {
	cs userpb.UserServiceClient
}

func NewClientService(cs userpb.UserServiceClient) *ClientService {
	return &ClientService{cs}
}

func (c *ClientService) CreateUser(ctx context.Context, user *userpb.User) (int, error) {
	res, err := c.cs.CreateUser(ctx, &userpb.CreateRequest{User: user})
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

func (c *ClientService) GetAllUser(ctx context.Context) ([]*userpb.User, error) {

	stream, err := c.cs.GetAllUser(ctx, &userpb.GetAllRequest{})
	if err != nil {
		return nil, fmt.Errorf("reading stream GetAllUser %v", err)
	}
	users := make([]*userpb.User, 0)
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

func (c *ClientService) DeleteUser(ctx context.Context, id int) error {
	_, err := c.cs.DeleteUser(ctx, &userpb.DeleteRequest{Id: int32(id)})
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientService) IssueToken(ctx context.Context, user *userpb.User) (string, error) {
	res, err := c.cs.IssueToken(ctx, &userpb.TokenRequest{User: user})
	if err != nil {
		return "", fmt.Errorf("error AuthUser %v", err)
	}
	token := res.Token
	return token.Token, nil
}

func (c *ClientService) AuthUser(ctx context.Context, token string) error {
	_, err := c.cs.AuthUser(ctx, &userpb.AuthRequest{
		Token: &userpb.Token{Token: token},
	})
	if err != nil {
		return fmt.Errorf("error ValidateToken %v", err)
	}
	return nil
}
