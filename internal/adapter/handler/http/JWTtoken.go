package http

import (
	"errors"
	"time"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTTokenService struct {
	secretKey  []byte
	expiration time.Duration
	logger     ports.LoggerPort
}

func NewJWTTokenService(secretKey string, durationStr string, logger ports.LoggerPort) *JWTTokenService {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		logger.Error("Invalid token duration, using default 24h", map[string]interface{}{
			"duration": durationStr,
			"error":    err.Error(),
		})
		duration = 24 * time.Hour
	}

	return &JWTTokenService{
		secretKey:  []byte(secretKey),
		expiration: duration,
		logger:     logger,
	}
}

func (j *JWTTokenService) CreateToken(user *domain.User) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		j.logger.Error("Failed to generate uuid", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
			"method":  "Create token",
		})
		return "", err
	}

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(j.expiration)

	claims := jwt.MapClaims{
		"id":      id.String(),
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"iat":     issuedAt.Unix(),
		"exp":     expiredAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTTokenService) VerifyToken(token string) (domain.TokenPayload, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		j.logger.Error("Failed to parse jwt", map[string]interface{}{
			"error":  err.Error(),
			"method": "VerifyToken",
		})
		return domain.TokenPayload{}, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		j.logger.Error("Failed claims from token", map[string]interface{}{
			"method": "VerifyToken",
		})
		return domain.TokenPayload{}, errors.New("Failed to verify")
	}

	idStr, ok := claims["id"].(string)
	if !ok {
		return domain.TokenPayload{}, errors.New("invalid id convert")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return domain.TokenPayload{}, errors.New("invalid parse id")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return domain.TokenPayload{}, errors.New("invalid user_id claims")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return domain.TokenPayload{}, errors.New("invalid parse user_id")
	}

	roleClaimed, ok := claims["role"].(string)
	if !ok {
		return domain.TokenPayload{}, errors.New("invalid role")
	}

	role := domain.UserRole(roleClaimed)
	if role != domain.Admin && role != domain.AppUser {
		j.logger.Warn("Invalid role in token", map[string]interface{}{
			"role":   roleClaimed,
			"method": "VerifyToken",
		})
		return domain.TokenPayload{}, errors.New("invalid role value")
	}

	payload := domain.TokenPayload{
		ID:     id,
		UserID: userID,
		Role:   role,
	}

	return payload, nil
}
