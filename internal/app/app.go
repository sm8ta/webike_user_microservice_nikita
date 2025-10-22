package app

import (
	"context"
	"time"
	grpcapp "webike_services/webike_User-microservice_Nikita/internal/app/grpc"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"
	"webike_services/webike_User-microservice_Nikita/internal/core/services"
	"webike_services/webike_User-microservice_Nikita/internal/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
	BikeClient *grpc.BikeClient
}

func New(
	ctx context.Context,
	log ports.LoggerPort,
	userService *services.UserService,
	grpcServerPort int,
	bikeServiceAddr string,
) (*App, error) {
	// Create gRPC server
	grpcServer := grpcapp.New(log, userService, grpcServerPort)

	// Create gRPC client, calls server
	bikeClient, err := grpc.NewBikeClient(
		ctx,
		log,
		bikeServiceAddr,
		5*time.Second, // timeout
		3,             // retries
	)
	if err != nil {
		log.Warn("Failed to connect to Bike Service", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return &App{
		GRPCServer: grpcServer,
		BikeClient: bikeClient,
	}, nil
}

// Run gRPC server
func (a *App) Run() {
	go a.GRPCServer.MustRun()
}

// Stop application
func (a *App) Stop() {
	a.GRPCServer.Stop()
}
