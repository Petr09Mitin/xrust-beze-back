package user_http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Petr09Mitin/xrust-beze-back/internal/middleware"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService user_service.UserService
}

func NewUserHandler(router *gin.Engine, userService user_service.UserService, authClient authpb.AuthServiceClient) {
	handler := &UserHandler{
		userService: userService,
	}

	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("/create", handler.Create)
		userGroup.GET("/:id", handler.GetByID)
		userGroup.GET("", handler.List)
		userGroup.GET("/match/:id", handler.FindMatchingUsers)
	}

	secure := router.Group("/api/v1/users")
	secure.Use(middleware.AuthMiddleware(authClient))
	{
		secure.PUT("/:id", handler.Update)
		secure.DELETE("/:id", handler.Delete)
	}
}

// ручка для тестов, регистрация в сервисе auth
func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword := "dummy"
	if err := h.userService.Create(ctx, &input, hashedPassword); err != nil {
		// Проверяем, является ли ошибка ошибкой валидации
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	// userIDCtx, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	// 	// в общем случае такие не должны допускаться сюда мидлварой
	// 	return
	// }
	// userIDStr, ok := userIDCtx.(string)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id format"})
	// 	return
	// }
	// paramID := c.Param("id")
	// if userIDStr != paramID {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another user's data"})
	// 	return
	// }
	// userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
	// 	return
	// }
	userObjectID, err := extractAndValidateUserID(c)
	if err != nil {
		return // JSON-ответ уже установлен в функции
	}

	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = userObjectID

	if err := h.userService.Update(ctx, &input); err != nil {
		if errors.Is(err, custom_errors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		// Проверяем, является ли ошибка ошибкой валидации
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	_, err := extractAndValidateUserID(c)
	if err != nil {
		return // JSON-ответ уже установлен в функции
	}

	id := c.Param("id")
	err = h.userService.Delete(ctx, id)
	if errors.Is(err, custom_errors.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Получение списка пользователей
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	users, err := h.userService.List(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Поиск подходящих пользователей
func (h *UserHandler) FindMatchingUsers(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	users, err := h.userService.FindMatchingUsers(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// ExtractAndValidateUserID извлекает user_id из контекста и проверяет соответствие параметру запроса
func extractAndValidateUserID(c *gin.Context) (primitive.ObjectID, error) {
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return primitive.NilObjectID, custom_errors.ErrMissingUserID
	}

	userIDStr, ok := userIDCtx.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id format"})
		return primitive.NilObjectID, custom_errors.ErrInvalidUserIDType
	}

	paramID := c.Param("id")
	if userIDStr != paramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another user's data"})
		return primitive.NilObjectID, custom_errors.ErrUserIDMismatch
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return primitive.NilObjectID, custom_errors.ErrInvalidUserID
	}

	return userObjectID, nil
}
