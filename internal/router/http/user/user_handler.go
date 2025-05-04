package user_http

import (
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"net/http"
	"strconv"
	"strings"

	"github.com/Petr09Mitin/xrust-beze-back/internal/middleware"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/httpparser"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
	middleware2 "github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService user_service.UserService
}

func NewUserHandler(router *gin.Engine, userService user_service.UserService, authClient authpb.AuthServiceClient) {
	handler := &UserHandler{
		userService: userService,
	}

	router.Use(middleware2.CORSMiddleware())
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.GET("/:id", handler.GetByID)
		userGroup.GET("", handler.List)
		userGroup.GET("/match/:id", handler.FindMatchingUsers)
		userGroup.GET("/by-name", handler.FindByUsername)
	}

	secure := router.Group("/api/v1/users")
	secure.Use(middleware.AuthMiddleware(authClient))
	{
		secure.PUT("/:id", handler.Update)
		secure.DELETE("/:id", handler.Delete)
		secure.POST("/review", handler.CreateReview)
		secure.PUT("/review/:id", handler.UpdateReview)
		secure.DELETE("/review/:id", handler.DeleteReview)
	}
}

func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	userObjectID, err := extractAndValidateUserID(c)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	var input user_model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	input.ID = userObjectID

	if err := input.Validate(); err != nil {
		if validationResp := validation.BuildValidationError(err); validationResp != nil {
			c.JSON(http.StatusBadRequest, validationResp)
			return
		}
		custom_errors.WriteHTTPError(c, err)
		return
	}

	if err := h.userService.Update(ctx, &input); err != nil {
		var profanityErr *custom_errors.ProfanityAggregateError
		if errors.As(err, &profanityErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                  profanityErr.Error(),
				"profanity_error_fields": profanityErr.Fields,
			})
			return
		}
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	_, err := extractAndValidateUserID(c)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	id := c.Param("id")
	if err = h.userService.Delete(ctx, id); err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (h *UserHandler) CreateReview(c *gin.Context) {
	ctx := c.Request.Context()
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userIDCtx.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input user_model.Review
	if err := c.ShouldBindJSON(&input); err != nil {
		custom_errors.WriteHTTPError(c, custom_errors.ErrBadRequest)
		return
	}
	input.UserIDBy = strings.TrimSpace(userIDStr)
	input.UserIDTo = strings.TrimSpace(input.UserIDTo)
	if input.UserIDBy == input.UserIDTo {
		custom_errors.WriteHTTPError(c, custom_errors.ErrCanNotSelfReview)
		return
	}
	if err := validation.Validate.Struct(input); err != nil {
		if validationResp := validation.BuildValidationError(err); validationResp != nil {
			c.JSON(http.StatusBadRequest, validationResp)
			return
		}
		custom_errors.WriteHTTPError(c, err)
		return
	}
	review, err := h.userService.CreateReview(ctx, &input)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusCreated, review)
}

func (h *UserHandler) UpdateReview(c *gin.Context) {
	ctx := c.Request.Context()
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userIDCtx.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	var input user_model.Review
	if err := c.ShouldBindJSON(&input); err != nil {
		custom_errors.WriteHTTPError(c, custom_errors.ErrBadRequest)
		return
	}
	input.ID = id
	review, err := h.userService.UpdateReview(ctx, userIDStr, &input)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, review)
}

func (h *UserHandler) DeleteReview(c *gin.Context) {
	ctx := c.Request.Context()
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userIDCtx.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if err := h.userService.DeleteReview(ctx, userIDStr, id); err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, gin.H{"message": "review deleted successfully"})
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
		custom_errors.WriteHTTPError(c, err)
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
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) FindByUsername(c *gin.Context) {
	username := strings.TrimSpace(c.Query("username"))
	userID := strings.TrimSpace(c.Query("user_id"))
	if username == "" || userID == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoUsername)
		return
	}
	limit, offset := httpparser.GetLimitAndOffset(c)
	users, err := h.userService.FindUsersByUsername(c.Request.Context(), userID, username, limit, offset)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// Извлекает user_id из контекста и проверяет соответствие параметру запроса
func extractAndValidateUserID(c *gin.Context) (bson.ObjectID, error) {
	userIDCtx, exists := c.Get("user_id")
	if !exists {
		return bson.NilObjectID, custom_errors.ErrMissingUserID
	}

	userIDStr, ok := userIDCtx.(string)
	if !ok {
		return bson.NilObjectID, custom_errors.ErrInvalidUserIDType
	}

	paramID := c.Param("id")
	if userIDStr != paramID {
		return bson.NilObjectID, custom_errors.ErrUserIDMismatch
	}

	userObjectID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return bson.NilObjectID, custom_errors.ErrInvalidUserID
	}

	return userObjectID, nil
}
