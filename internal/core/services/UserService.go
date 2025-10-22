package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/ports"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo     ports.UserRepository
	logger   ports.LoggerPort
	validate *validator.Validate
	cache    ports.CachePort
}

func NewUserService(
	repo ports.UserRepository,
	logger ports.LoggerPort,
	validate *validator.Validate,
	cache ports.CachePort,

) *UserService {
	return &UserService{
		repo:     repo,
		logger:   logger,
		validate: validate,
		cache:    cache,
	}
}

func (us *UserService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := us.validateUser(user); err != nil {
		us.logger.Error("Validation failed", map[string]interface{}{
			"error":  err.Error(),
			"method": "Register",
		})
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		us.logger.Error("Error during hashing", map[string]interface{}{
			"error":  err.Error(),
			"method": "Register",
		})
		return nil, err
	}

	user.Password = string(hashedPassword)

	user, err = us.repo.CreateUser(ctx, user)
	if err != nil {
		us.logger.Error("Failed to create user in database", map[string]interface{}{
			"error":  err.Error(),
			"method": "Register",
		})
		return nil, err
	}
	return user, nil
}

func (us *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		us.logger.Error("Invalid UUID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	// Tries to take from cache
	cacheKey := fmt.Sprintf("user:%s", id)
	cachedData, err := us.cache.Get(cacheKey)
	if err == nil {
		var cachedUser domain.User
		if err := json.Unmarshal(cachedData, &cachedUser); err == nil {
			us.logger.Info("User found in cache", map[string]interface{}{
				"id": id,
			})
			return &cachedUser, nil
		}
	}

	// Going to db
	user, err := us.repo.GetUserByID(ctx, userID)
	if err != nil {
		us.logger.Error("Failed to get user", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, err
	}

	// Caching for 15 min
	userData, err := json.Marshal(user)
	if err != nil {
		us.logger.Warn("Failed to marshal user for cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
	} else {
		if err := us.cache.Set(cacheKey, userData, 15*time.Minute); err != nil {
			us.logger.Warn("Failed to cache user", map[string]interface{}{
				"error": err.Error(),
				"id":    id,
			})
		}
	}

	return user, nil
}

func (us *UserService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			us.logger.Error("Error during hashing", map[string]interface{}{
				"error":  err.Error(),
				"method": "UpdateUser",
			})
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	updatedUser, err := us.repo.UpdateUser(ctx, user)
	if err != nil {
		us.logger.Error("Failed to update user", map[string]interface{}{
			"id":    user.ID,
			"error": err.Error(),
		})
		return nil, err
	}

	cacheKey := fmt.Sprintf("user:%s", user.ID.String())
	if err := us.cache.Delete(cacheKey); err != nil {
		us.logger.Warn("Failed to invalidate user cache", map[string]interface{}{
			"error": err.Error(),
			"id":    user.ID.String(),
		})
	}

	emailCacheKey := fmt.Sprintf("user_email:%s", updatedUser.Email)
	if err := us.cache.Delete(emailCacheKey); err != nil {
		us.logger.Warn("Failed to invalidate user email cache", map[string]interface{}{
			"error": err.Error(),
			"email": updatedUser.Email,
		})
	}

	return updatedUser, nil
}

func (us *UserService) DeleteUser(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		us.logger.Error("Invalid UUID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return fmt.Errorf("invalid ID format: %w", err)
	}

	user, err := us.repo.GetUserByID(ctx, userID)
	if err != nil {
		us.logger.Error("Failed to get user before deletion", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	if err := us.repo.DeleteUser(ctx, userID); err != nil {
		us.logger.Error("Failed to delete user", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	cacheKey := fmt.Sprintf("user:%s", id)
	if err := us.cache.Delete(cacheKey); err != nil {
		us.logger.Warn("Failed to invalidate user cache", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
	}

	emailCacheKey := fmt.Sprintf("user_email:%s", user.Email)
	if err := us.cache.Delete(emailCacheKey); err != nil {
		us.logger.Warn("Failed to invalidate user email cache", map[string]interface{}{
			"error": err.Error(),
			"email": user.Email,
		})
	}

	us.logger.Info("User deleted", map[string]interface{}{
		"id": id,
	})
	return nil
}

func (us *UserService) validateUser(user *domain.User) error {
	if err := us.validate.Struct(user); err != nil {
		return fmt.Errorf("validation failed: %s", err.Error())
	}

	date, err := time.Parse("2006-01-02", user.DateOfBirth)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	age := time.Now().Year() - date.Year()
	if time.Now().Before(date.AddDate(age, 0, 0)) {
		age--
	}

	if age < 6 {
		return fmt.Errorf("user must be at least 6 years old")
	}

	return nil
}
