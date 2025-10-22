package http

import (
	"webike_services/webike_User-microservice_Nikita/internal/core/domain"

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
