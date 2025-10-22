package http

import (
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"

	"github.com/gin-gonic/gin"
)

func getAuthPayload(ctx *gin.Context, key string) (*domain.TokenPayload, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return nil, false
	}
	payload, ok := value.(*domain.TokenPayload)
	if !ok {
		return nil, false
	}
	return payload, true
}
