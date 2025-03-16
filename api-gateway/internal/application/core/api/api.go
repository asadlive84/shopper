package api

import (
	"context"
	"fmt"
	"time"

	pb "github.com/asadlive84/shopper-proto/golang/user"
	"github.com/asadlive84/shopper/api-gateway/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type APIService struct {
	grpcClient ports.UserGRPCClientPort
	jwtSecret  string
	logger     *zap.Logger
}

func NewAPIService(grpcClient ports.UserGRPCClientPort, jwtSecret string, logger *zap.Logger) *APIService {
	return &APIService{
		grpcClient: grpcClient,
		jwtSecret:  jwtSecret,
		logger:     logger,
	}
}

func (s *APIService) GetUser(ctx context.Context, id string) (*pb.User, error) {
	req := &pb.GetUserRequest{Id: id}
	res, err := s.grpcClient.GetUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetUser(), nil
}

func (s *APIService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	res, err := s.grpcClient.CreateUser(ctx, req)
	if err != nil {
		return nil, s.handleGRPCError(err, "CreateUser", zap.String("email", req.Email))
	}
	if res == nil || res.User == nil {
		return nil, fmt.Errorf("empty response received from server")
	}
	return res.User, nil
}

func (s *APIService) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	res, err := s.grpcClient.AuthenticateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": res.GetUser().GetId(),
		"email":   res.GetUser().GetEmail(),
		"role":    res.GetUser().GetRole().String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	fmt.Println("========================================")
	fmt.Printf("token %+v\n", token)
	fmt.Printf("res.GetUser() %+v\n", res.GetUser())
	fmt.Println("========================================")

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}
	return &pb.AuthenticateUserResponse{
		Token: tokenString,
		User:  res.GetUser(),
	}, nil
}

func (s *APIService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return s.grpcClient.ListUsers(ctx, req)
}

func (s *APIService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	res, err := s.grpcClient.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetUser(), nil
}
