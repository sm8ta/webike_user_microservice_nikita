package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}
type UserInfo struct {
	ID    uuid.UUID       `json:"id"`
	Email string          `json:"email"`
	Name  string          `json:"name"`
	Role  domain.UserRole `json:"role"`
}
type AuthHandler struct {
	authService ports.AuthService
	logger      ports.LoggerPort
	metrics     ports.MetricsPort
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

func NewAuthHandler(
	authService ports.AuthService,
	logger ports.LoggerPort,
	metrics ports.MetricsPort,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		metrics:     metrics,
	}
}

// @Summary Авторизация пользователя
// @Description Вход в систему по email и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse "Успешная авторизация"
// @Failure 400 {object} errorResponse "Неверный запрос"
// @Failure 401 {object} errorResponse "Неверные учетные данные"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {

	// Getting metrics
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed JSON parse in login", map[string]interface{}{
			"error": err.Error(),
		})
		newErrorResponse(c, http.StatusBadRequest, "Invalid request")
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Info("Login failed", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})
		newErrorResponse(c, http.StatusUnauthorized, "Invalid data")
		return
	}

	h.logger.Info("User logged in successfully", map[string]interface{}{
		"email":   req.Email,
		"user_id": user.ID,
	})

	response := LoginResponse{
		Token: token,
		User: UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}

	c.JSON(http.StatusOK, response)
}
