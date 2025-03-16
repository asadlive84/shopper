package grpc

import (
	"net"

	pb "github.com/asadlive84/shopper-proto/golang/user"
	"github.com/asadlive84/shopper/user-svc/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	api    ports.APIPort
	logger *zap.Logger
}

func NewUserServer(api ports.APIPort, logger *zap.Logger) *UserServer {
	return &UserServer{api: api, logger: logger}
}

func StartGRPCServer(port string, api ports.APIPort, logger *zap.Logger, server *grpc.Server) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	userServer := NewUserServer(api, logger)
	pb.RegisterUserServiceServer(server, userServer)

	logger.Info("gRPC server listening", zap.String("port", port))
	return server.Serve(lis)
}
