package http

import (
	"net/http"
	"time"

	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/domain"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/ports"
	"github.com/sm8ta/webike_user_microservice_nikita/internal/core/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService  *services.UserService
	logger       ports.LoggerPort
	tokenService *JWTTokenService
	metrics      ports.MetricsPort
}

type UserRequest struct {
	Name        string `json:"name" binding:"required" example:"Иван Иванов"`
	DateOfBirth string `json:"date_of_birth" binding:"required" example:"1990-01-01"`
	Email       string `json:"email" binding:"required" example:"ivan@example.com"`
	Password    string `json:"password" binding:"required" example:"password123"`
}

type UpdateUser struct {
	Name        *string `json:"name,omitempty" example:"Новое имя"`
	DateOfBirth *string `json:"date_of_birth,omitempty" example:"1990-01-01"`
	Email       *string `json:"email,omitempty" example:"new@example.com"`
	Password    *string `json:"password,omitempty" example:"newpassword123"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUserResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DateOfBirth string    `json:"date_of_birth"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateUserResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DateOfBirth string    `json:"date_of_birth"`
	Role        string    `json:"role"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}

func NewUserHandler(
	userService *services.UserService,
	logger ports.LoggerPort,
	tokenService *JWTTokenService,
	metrics ports.MetricsPort,
) *UserHandler {
	return &UserHandler{
		userService:  userService,
		logger:       logger,
		tokenService: tokenService,
		metrics:      metrics,
	}
}

// @Summary Регистрация пользователя
// @Description Создание нового пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param request body UserRequest true "Данные пользователя"
// @Success 201 {object} RegisterResponse "Пользователь создан"
// @Failure 400 {object} errorResponse "Неверный запрос"
// @Failure 409 {object} errorResponse "Email уже существует"
// @Router /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	var req UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed JSON parse in registration", map[string]interface{}{
			"error": err.Error(),
		})
		newErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	ctx := c.Request.Context()

	user := &domain.User{
		Name:        req.Name,
		DateOfBirth: req.DateOfBirth,
		Email:       req.Email,
		Password:    req.Password,
		Role:        domain.AppUser,
	}

	createdUser, err := h.userService.Register(ctx, user)
	if err != nil {
		if err.Error() == "email already exists" {
			h.logger.Info("Registration failed: duplicate email", map[string]interface{}{
				"email": req.Email,
			})
			newErrorResponse(c, http.StatusConflict, "Email already registered")
			return
		}

		h.logger.Error("Failed to register user", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		newErrorResponse(c, http.StatusInternalServerError, "Registration failed")
		return
	}

	token, err := h.tokenService.CreateToken(createdUser)
	if err != nil {
		h.logger.Error("Failed to create token", map[string]interface{}{
			"error": err.Error(),
		})
		newErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.logger.Info("User created successfully", map[string]interface{}{
		"email":   req.Email,
		"user_id": createdUser.ID,
	})

	response := RegisterResponse{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		Token:     token,
		CreatedAt: createdUser.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Получить пользователя
// @Description Получение информации о пользователе по ID
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID юзера" example:"jdk2-fsjmk-daslkdo2-321md-jsnlaljdn"
// @Success 200 {object} GetUserResponse "Пользователь найден"
// @Failure 401 {object} errorResponse "Не авторизован"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Failure 404 {object} errorResponse "Пользователь не найден"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	userID := c.Param("id")

	payload, exists := getAuthPayload(c, "authorization_payload")
	if !exists {
		h.logger.Warn("Unauthorized access attempt to GetUser", map[string]interface{}{
			"requested_user_id": userID,
			"ip":                c.ClientIP(),
		})
		newErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if payload.Role != domain.Admin && payload.UserID.String() != userID {
		h.logger.Warn("Access denied to user profile", map[string]interface{}{
			"requester_id": payload.UserID.String(),
			"requested_id": userID,
			"role":         payload.Role,
		})
		newErrorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		newErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}
	response := GetUserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		DateOfBirth: user.DateOfBirth,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Обновить пользователя
// @Description Обновление данных пользователя
// @Tags users
// @Security BearerAuth
// @Param id path string true "ID юзера" example:"jdk2-fsjmk-daslkdo2-321md-jsnlaljdn"
// @Param request body UpdateUser true "Данные для обновления"
// @Success 200 {object} UpdateUserResponse "Пользователь обновлен"
// @Failure 400 {object} errorResponse "Неверный запрос"
// @Failure 401 {object} errorResponse "Не авторизован"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	userID := c.Param("id")

	payload, exists := getAuthPayload(c, "authorization_payload")
	if !exists {
		h.logger.Warn("Unauthorized access attempt to UpdateUser", map[string]interface{}{
			"requested_user_id": userID,
			"ip":                c.ClientIP(),
		})
		newErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if payload.Role != domain.Admin && payload.UserID.String() != userID {
		h.logger.Warn("Access denied to update user", map[string]interface{}{
			"requester_id": payload.UserID.String(),
			"requested_id": userID,
			"role":         payload.Role,
		})
		newErrorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	var req UpdateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed JSON parse in update user", map[string]interface{}{
			"error": err.Error(),
		})
		newErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Invalid user ID format", map[string]interface{}{
			"user_id": userID,
		})
		newErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user := &domain.User{ID: parsedID}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = *req.DateOfBirth
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		user.Password = *req.Password
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		if err.Error() == "email already exists" {
			h.logger.Info("Update failed: duplicate email", map[string]interface{}{
				"email":   user.Email,
				"user_id": userID,
			})
			newErrorResponse(c, http.StatusConflict, "Email already exists")
			return
		}
		h.logger.Error("Failed to update user", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		newErrorResponse(c, http.StatusInternalServerError, "Update failed")
		return
	}

	h.logger.Info("User updated successfully", map[string]interface{}{
		"user_id": userID,
	})

	response := UpdateUserResponse{
		ID:          updatedUser.ID,
		Name:        updatedUser.Name,
		Email:       updatedUser.Email,
		DateOfBirth: updatedUser.DateOfBirth,
		Role:        string(updatedUser.Role),
		UpdatedAt:   updatedUser.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Удалить пользователя
// @Description Удаление пользователя
// @Tags users
// @Security BearerAuth
// @Param id path string true "ID юзера" example:"jdk2-fsjmk-daslkdo2-321md-jsnlaljdn"
// @Success 200 {object} DeleteUserResponse "Пользователь удален"
// @Failure 401 {object} errorResponse "Не авторизован"
// @Failure 403 {object} errorResponse "Доступ запрещен"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	userID := c.Param("id")

	payload, exists := getAuthPayload(c, "authorization_payload")
	if !exists {
		h.logger.Warn("Unauthorized access attempt to DeleteUser", map[string]interface{}{
			"requested_user_id": userID,
			"ip":                c.ClientIP(),
		})
		newErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if payload.Role != domain.Admin && payload.UserID.String() != userID {
		h.logger.Warn("Access denied to delete user", map[string]interface{}{
			"requester_id": payload.UserID.String(),
			"requested_id": userID,
		})
		newErrorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete user", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		newErrorResponse(c, http.StatusInternalServerError, "Delete failed")
		return
	}

	h.logger.Info("User deleted successfully", map[string]interface{}{
		"user_id": userID,
	})

	c.JSON(http.StatusOK, DeleteUserResponse{
		Message: "User deleted successfully",
	})
}

/*

func (h *UserHandler) GetUserWithBikes(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordMetrics(c, start)
	}()

	userID := c.Param("id")

	payload, exists := getAuthPayload(c, "authorization_payload")
	if !exists {
		h.logger.Warn("Unauthorized access attempt to GetUserWithBikes", map[string]interface{}{
			"requested_user_id": userID,
			"ip":                c.ClientIP(),
		})
		newErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if payload.Role != domain.Admin && payload.UserID.String() != userID {
		h.logger.Warn("Access denied to user profile", map[string]interface{}{
			"requester_id": payload.UserID.String(),
			"requested_id": userID,
			"role":         payload.Role,
		})
		newErrorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Получаем юзера
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		newErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// Вызов Bike-сервиса через HTTP клиент
	params := bikes.NewGetBikesMyParams()
	params.Context = c.Request.Context()

	// Подготовка авторизации
	authHeader := c.GetHeader("Authorization")
	var authInfo runtime.ClientAuthInfoWriter
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		authInfo = httptransport.BearerToken(token)
	}

	bikesResp, err := h.bikeClient.Bikes.GetBikesMy(params, authInfo)

	var bikesList []*models.DomainBike
	if err != nil {
		h.logger.Warn("Failed to get bikes from Bike service", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		bikesList = []*models.DomainBike{}
	} else {
		bikesList = bikesResp.Payload.Data
	}

	response := UserWithBikesResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		Email:       user.Email,
		DateOfBirth: user.DateOfBirth,
		Bikes:       bikesList,
	}

	newSuccessResponse(c, http.StatusOK, "User with bikes found", response)
}

*/
