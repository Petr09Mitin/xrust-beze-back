package user_http

import (
	"net/http"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	middleware2 "github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"

	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	"github.com/gin-gonic/gin"
)

type SkillHandler struct {
	skillService user_service.SkillService
}

func NewSkillHandler(router *gin.Engine, service user_service.SkillService) {
	handler := &SkillHandler{
		skillService: service,
	}

	router.Use(middleware2.CORSMiddleware())
	skillGroup := router.Group("/api/v1/skills")
	{
		skillGroup.GET("", handler.GetAllSkills)
		skillGroup.GET("/:category", handler.GetSkillsByCategory)
		skillGroup.GET("/categories", handler.GetAllCategories)
	}
}

func (h *SkillHandler) GetAllSkills(c *gin.Context) {
	ctx := c.Request.Context()
	allSkills, err := h.skillService.GetAll(ctx)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, allSkills)
}

func (h *SkillHandler) GetSkillsByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")
	skills, err := h.skillService.GetByCategory(ctx, category)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	response := gin.H{category: skills}
	c.JSON(http.StatusOK, response)
}

func (h *SkillHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.skillService.GetAllCategories(c.Request.Context())
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
