package http

import (
	"net/http"
	"strconv"

	errs "github.com/Petr09Mitin/xrust-beze-back/internal/models/errs"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService user_service.UserService
}

func NewUserHandler(router *gin.Engine, userService user_service.UserService) {
	handler := &UserHandler{
		userService: userService,
	}

	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("", handler.Create)
		userGroup.GET("/:id", handler.GetByID)
		userGroup.PUT("/:id", handler.Update)
		userGroup.DELETE("/:id", handler.Delete)
		userGroup.GET("", handler.List)
		userGroup.GET("/match/:id", handler.FindMatchingUsers)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Create(ctx, &input); err != nil {
		// Проверяем, является ли ошибка ошибкой валидации
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// добавить в слой services проверку user id

	ctx := c.Request.Context()
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ID = objectID

	if err := h.userService.Update(ctx, &input); err != nil {
		if err == errs.ErrNotFound {
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

	// добавить в слой services проверку user id

	ctx := c.Request.Context()
	id := c.Param("id")

	err := h.userService.Delete(ctx, id)

	if err == errs.ErrNotFound {
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
