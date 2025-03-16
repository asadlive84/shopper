package ports

import (
    "context"
    pb "github.com/asadlive84/shopper-proto/golang/user"
)

type APIPort interface {
    GetUser(ctx context.Context, id string) (*pb.User, error)
    CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error)
    AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error)
    ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
    UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error)
    
}