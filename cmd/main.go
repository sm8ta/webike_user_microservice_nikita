package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/sm8ta/webike_user_microservice_nikita/docs"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/handler/http"
	handlers "github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/handler/http"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/logger"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/prometheus"
	redis "github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/redis"

	redisClient "github.com/redis/go-redis/v9"

	"github.com/sm8ta/webike_user_microservice_nikita/internal/adapter/postgres/repository"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/config"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/services"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

// @title User Microservice API
// @version 1.1
// @description API для управления пользователями

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Loading environment
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	// Set redis
	redisConn := redisClient.NewClient(&redisClient.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	ctx := context.Background()
	if _, err := redisConn.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	// Set logger
	loggerAdapter := logger.NewLoggerAdapter(cfg.App.Env)
	loggerAdapter.Info("Starting the application", map[string]interface{}{
		"app": cfg.App.Name,
		"env": cfg.App.Env,
	})

	// Connect DB
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	// Migrate DB
	if err := goose.Up(db, "./internal/adapter/postgres/migrations"); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	// Cache
	cacheAdapter := redis.NewRedisAdapter(redisConn)

	// Validate
	validate := validator.New()

	// Observability
	metrics := prometheus.NewPrometheusAdapter()

	// User
	userRepo := repository.NewUserRepository(db)
	tokenService := handlers.NewJWTTokenService(cfg.Token.Secret, cfg.Token.Duration, loggerAdapter)
	authService := services.NewAuthService(userRepo, tokenService, loggerAdapter, cacheAdapter)
	authHandler := handlers.NewAuthHandler(authService, loggerAdapter, metrics)
	userService := services.NewUserService(userRepo, loggerAdapter, validate, cacheAdapter)

	userHandler := handlers.NewUserHandler(userService, loggerAdapter, tokenService, metrics)

	// Init router
	router, err := http.NewRouter(
		cfg.HTTP,
		tokenService,
		userHandler,
		authHandler,
	)
	if err != nil {
		log.Fatal("Error initializing router:", err)
	}

	go func() {
		listenAddr := fmt.Sprintf("%s:%s", cfg.HTTP.URL, cfg.HTTP.Port)
		loggerAdapter.Info("Starting the HTTP server", map[string]interface{}{
			"addr": listenAddr,
		})

		if err := router.Serve(listenAddr); err != nil {
			log.Fatal("Error starting the HTTP server:", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	loggerAdapter.Info("Application is running", nil)

	<-stop

	loggerAdapter.Info("Shut", nil)
	loggerAdapter.Info("Application stopped", nil)
}
