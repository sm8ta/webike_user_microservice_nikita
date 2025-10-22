package http

import (
	"net/http"
	"strings"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/ports"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationType       = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(token ports.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Auth header required",
			})
			c.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Auth fields required",
			})
			c.Abort()
			return
		}

		currentAuthorizationType := strings.ToLower(fields[0])
		if currentAuthorizationType != authorizationType {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authorizated",
			})
			c.Abort()
			return
		}

		accessToken := fields[1]
		payload, err := token.VerifyToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(authorizationPayloadKey, &payload)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, ok := getAuthPayload(ctx, authorizationPayloadKey)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization required",
			})
			ctx.Abort()
			return
		}

		if payload.Role != domain.Admin {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
