package http

import (
	"net/http"
	"strconv"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler представляет HTTP обработчик для пользователей
type UserHandler struct {
	userService user_model.Service
}

// NewUserHandler создает новый обработчик пользователей
func NewUserHandler(router *gin.Engine, userService user_model.Service) {
	handler := &UserHandler{
		userService: userService,
	}

	userGroup := router.Group("/api/users")
	{
		userGroup.POST("", handler.Create)
		userGroup.GET("/:id", handler.GetByID)
		userGroup.PUT("/:id", handler.Update)
		userGroup.DELETE("/:id", handler.Delete)
		userGroup.GET("", handler.List)
		userGroup.GET("/match/:id", handler.FindMatchingUsers)
	}
}

// Create обрабатывает создание нового пользователя
func (h *UserHandler) Create(c *gin.Context) {
	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Create(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// GetByID обрабатывает получение пользователя по ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update обрабатывает обновление пользователя
func (h *UserHandler) Update(c *gin.Context) {
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

	if err := h.userService.Update(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

// Delete обрабатывает удаление пользователя
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.userService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
}

// List обрабатывает получение списка пользователей
func (h *UserHandler) List(c *gin.Context) {
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

	users, err := h.userService.List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// FindMatchingUsers обрабатывает поиск подходящих пользователей
func (h *UserHandler) FindMatchingUsers(c *gin.Context) {
	id := c.Param("id")

	users, err := h.userService.FindMatchingUsers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
} 