package http

import (
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error"`
}

type successResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message,omitempty" example:"Success message"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{
		Success: false,
		Message: message,
	})
}

func newSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, successResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
