package grpc

import (
	"context"
	"log"

	pb "github.com/asadlive84/shopper-proto/golang/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGRPCClient struct {
	client pb.UserServiceClient
}

func NewUserGRPCClient(addr string) *UserGRPCClient {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	return &UserGRPCClient{client: pb.NewUserServiceClient(conn)}
}

func (c *UserGRPCClient) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return c.client.GetUser(ctx, req)
}

func (c *UserGRPCClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return c.client.CreateUser(ctx, req)
}

func (c *UserGRPCClient) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	return c.client.AuthenticateUser(ctx, req)
}

func (c *UserGRPCClient) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return c.client.ListUsers(ctx, req)
}

func (c *UserGRPCClient) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return c.client.UpdateUser(ctx, req)
}
