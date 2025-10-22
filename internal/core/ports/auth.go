package ports

import (
	"context"
	"webike_services/webike_User-microservice_Nikita/internal/core/domain"
)

type TokenService interface {
	CreateToken(user *domain.User) (string, error)
	VerifyToken(token string) (domain.TokenPayload, error)
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
}
