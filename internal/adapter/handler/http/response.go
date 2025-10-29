package http

import (
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{
		Success: false,
		Message: message,
	})
}
