package http

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, allSkills)
}

func (h *SkillHandler) GetSkillsByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")
	skills, err := h.skillService.GetByCategory(ctx, category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	response := gin.H{category: skills}
	c.JSON(http.StatusOK, response)
}

func (h *SkillHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.skillService.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
