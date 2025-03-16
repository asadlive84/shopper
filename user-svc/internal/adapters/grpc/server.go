package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/asadlive84/shopper-proto/golang/user"
	"github.com/asadlive84/shopper/user-svc/internal/application/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate input fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		errMsg := "Invalid input: Name, Email, and Password are required fields"
		s.logger.Error(errMsg)
		return nil, status.Errorf(codes.InvalidArgument, errMsg)
	}

	// Create a new user object
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call the service layer to create the user
	createdUser, err := s.api.CreateUser(ctx, user)
	if err != nil {
		// Handle specific error cases
		var errMsg string
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			errMsg = fmt.Sprintf("User with email '%s' already exists", req.Email)
			s.logger.Error(errMsg, zap.Error(err))
			return nil, status.Errorf(codes.AlreadyExists, errMsg)

		case errors.Is(err, domain.ErrDatabaseError):
			errMsg = "Internal server error while creating the user"
			s.logger.Error(errMsg, zap.Error(err))
			return nil, status.Errorf(codes.AlreadyExists, errMsg)

		default:
			errMsg = "An unexpected error occurred while creating the user"
			s.logger.Error(errMsg, zap.Error(err))
			return nil, status.Errorf(codes.AlreadyExists, errMsg)
		}
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:       createdUser.ID.String(),
			Name:     createdUser.Name,
			Email:    createdUser.Email,
		},
	}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {

	user, err := s.api.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, err

	}
	return &pb.GetUserResponse{
		User: &pb.User{
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *UserServer) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	user, err := s.api.AuthenticateUser(ctx, &domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return nil, err

	}

	return &pb.AuthenticateUserResponse{
		User: &pb.User{
			Name:  user.User.Name,
			Email: user.User.Email,
		},
	}, nil
}
