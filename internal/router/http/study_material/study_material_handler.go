package study_material_http

import (
	"net/http"

	"github.com/Petr09Mitin/xrust-beze-back/internal/middleware"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
)

type StudyMaterialHandler struct {
	studyMaterialService study_material_service.StudyMaterialAPIService
}

func NewStudyMaterialHandler(router *gin.Engine, studyMaterialService study_material_service.StudyMaterialAPIService, authClient authpb.AuthServiceClient) {
	handler := &StudyMaterialHandler{
		studyMaterialService: studyMaterialService,
	}

	materialGroup := router.Group("/api/v1/study-materials")
	{
		materialGroup.GET("/:id", handler.GetByID)
		materialGroup.GET("/by-tags", handler.GetByTags)
		materialGroup.GET("", handler.GetByName)
		materialGroup.GET("/by-author-id/:author_id", handler.GetByAuthorID)
	}

	secure := router.Group("/api/v1/study-materials")
	secure.Use(middleware.AuthMiddleware(authClient))
	{
		secure.DELETE("/:id", handler.Delete)
	}
}

// Получение учебного материала по ID
func (h *StudyMaterialHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMaterialID)
		return
	}
	material, err := h.studyMaterialService.GetByID(c.Request.Context(), id)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, material)
}

// Поиск учебных материалов по тегам (по принципу "хотя бы один")
func (h *StudyMaterialHandler) GetByTags(c *gin.Context) {
	// tag := c.Query("tag")
	// 	if tag == "" {
	tags := c.QueryArray("tag") // получаем все ?tag=... из запроса
	if len(tags) == 0 {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMaterialTag)
		return
	}
	materials, err := h.studyMaterialService.GetByTags(c.Request.Context(), tags)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"study_materials": materials})
}

// Поиск учебных материалов по названию
func (h *StudyMaterialHandler) GetByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMaterialName)
		return
	}

	materials, err := h.studyMaterialService.GetByName(c.Request.Context(), name)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"study_materials": materials})
}

// Получение всех материалов одного автора
func (h *StudyMaterialHandler) GetByAuthorID(c *gin.Context) {
	authorID := c.Param("author_id")
	if authorID == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMaterialAuthorID)
		return
	}

	materials, err := h.studyMaterialService.GetByAuthorID(c.Request.Context(), authorID)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"study_materials": materials})
}

// Удаление учебного материала (только для автора)
func (h *StudyMaterialHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMaterialID)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		custom_errors.WriteHTTPError(c, custom_errors.ErrMissingUserID)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		custom_errors.WriteHTTPError(c, custom_errors.ErrInvalidUserIDType)
		return
	}

	err := h.studyMaterialService.Delete(c.Request.Context(), id, userIDStr)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study material deleted successfully"})
}
