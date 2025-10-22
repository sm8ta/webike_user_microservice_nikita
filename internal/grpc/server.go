package grpc

import (
	"context"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"
	"webike_services/webike_User-microservice_Nikita/internal/core/services"

	webikev1 "github.com/sm8ta/grpc_webike/gen/go/webike"
	"google.golang.org/grpc"
)

type serverAPI struct {
	webikev1.UnimplementedUserServiceServer
	userService *services.UserService
	log         ports.LoggerPort
}

func Register(
	gRPCServer *grpc.Server,
	userService *services.UserService,
	log ports.LoggerPort,
) {
	webikev1.RegisterUserServiceServer(gRPCServer, &serverAPI{
		userService: userService,
		log:         log,
	})
}

func (s *serverAPI) GetUser(ctx context.Context, req *webikev1.GetUserRequest) (*webikev1.GetUserResponse, error) {
	userID := req.GetUserId()

	user, err := s.userService.GetUser(ctx, req.GetUserId())
	if err != nil {
		s.log.WarnGRPC(ctx, "User not found", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return &webikev1.GetUserResponse{Exists: false}, nil
	}
	return &webikev1.GetUserResponse{
		Exists: true,
		UserId: user.ID.String(),
		Name:   user.Name,
	}, nil
}
