package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"webike_services/webike_User-microservice_Nikita/internal/core/domain"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     ports.UserRepository
	tokenService ports.TokenService
	logger       ports.LoggerPort
	cache        ports.CachePort
}

func NewAuthService(
	userRepo ports.UserRepository,
	tokenService ports.TokenService,
	logger ports.LoggerPort,
	cache ports.CachePort,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
		logger:       logger,
		cache:        cache,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	cacheKey := fmt.Sprintf("user_email:%s", email)
	cachedData, err := s.cache.Get(cacheKey)
	var user *domain.User

	if err == nil {
		var cachedUser domain.User
		if err := json.Unmarshal(cachedData, &cachedUser); err == nil {
			user = &cachedUser
			s.logger.Info("User found in email cache", map[string]interface{}{
				"email": email,
			})
		}
	}

	if user == nil {
		user, err = s.userRepo.GetUserByEmail(ctx, email)
		if err != nil {
			s.logger.Error("Failed to get user by email", map[string]interface{}{
				"email": email,
				"error": err.Error(),
			})
			return "", nil, errors.New("invalid credentials")
		}

		if user == nil {
			return "", nil, errors.New("invalid credentials")
		}

		userData, err := json.Marshal(user)
		if err != nil {
			s.logger.Warn("Failed to marshal user for email cache", map[string]interface{}{
				"error": err.Error(),
				"email": email,
			})
		} else {
			if err := s.cache.Set(cacheKey, userData, 10*time.Minute); err != nil {
				s.logger.Warn("Failed to cache user by email", map[string]interface{}{
					"error": err.Error(),
					"email": email,
				})
			}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Info("Invalid password attempt", map[string]interface{}{
			"email": email,
		})
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.tokenService.CreateToken(user)
	if err != nil {
		s.logger.Error("Failed to create token", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
		return "", nil, err
	}

	userResponse := *user
	userResponse.Password = ""
	return token, &userResponse, nil
}
