package ports

import (
	"context"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
)

type TokenService interface {
	CreateToken(user *domain.User) (string, error)
	VerifyToken(token string) (domain.TokenPayload, error)
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
}
