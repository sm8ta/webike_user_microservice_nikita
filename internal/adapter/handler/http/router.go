package http

import (
	"strings"
	"webike_services/webike_User-microservice_Nikita/internal/config"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	*gin.Engine
}

func NewRouter(
	config *config.HTTP,
	tokenService ports.TokenService,
	userHandler *UserHandler,
	authHandler *AuthHandler,
) (*Router, error) {
	if config.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// CORS
	ginConfig := cors.DefaultConfig()
	allowedOrigins := config.AllowedOrigins
	originsList := strings.Split(allowedOrigins, ",")
	ginConfig.AllowOrigins = originsList

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), cors.New(ginConfig))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers without auth
	router.POST("/register", userHandler.RegisterUser)
	router.POST("/login", authHandler.Login)

	// Routers with auth
	users := router.Group("/users")
	users.Use(AuthMiddleware(tokenService))
	{
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.GET("/:id/with-bikes", userHandler.GetUserWithBikes)
		users.DELETE("/:id", userHandler.DeleteUser)
	}

	return &Router{
		Engine: router,
	}, nil
}

// Starts the HTTP server
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}
