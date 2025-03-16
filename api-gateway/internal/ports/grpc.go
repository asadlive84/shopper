package ports

import (
    "context"
    pb "github.com/asadlive84/shopper-proto/golang/user"
)

type UserGRPCClientPort interface {
    GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error)
    CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error)
    AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error)
    ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
    UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
}