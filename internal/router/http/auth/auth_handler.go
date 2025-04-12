package auth_http

import (
	"net/http"
	"strings"

	auth_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/auth"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService auth.AuthService
	config      *config.Auth
}

func NewAuthHandler(router *gin.Engine, authService auth.AuthService, config *config.Auth) {
	handler := &AuthHandler{
		authService: authService,
		config:      config,
	}

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.GET("/validate", handler.Validate)
		// auth.GET("/test-connection", handler.TestConnection)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req auth_model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		if validationResp := validation.BuildValidationError(err); validationResp != nil {
			c.JSON(http.StatusBadRequest, validationResp)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		// var profanityErr *custom_errors.ProfanityAggregateError
		// if errors.As(err, &profanityErr) {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error":                  profanityErr.Error(),
		// 		"profanity_error_fields": profanityErr.Fields,
		// 	})
		// 	return
		// }
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	session, err := h.authService.CreateSession(c.Request.Context(), user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	domain, secure := h.resolveCookieSettings(c)
	c.SetCookie(
		h.config.Cookie.Name,
		session.ID,
		h.config.Cookie.MaxAge,
		"/",
		domain,
		secure,
		h.config.Cookie.HttpOnly,
	)

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req auth_model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, user, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	domain, secure := h.resolveCookieSettings(c)
	c.SetCookie(
		h.config.Cookie.Name,
		session.ID,
		h.config.Cookie.MaxAge,
		"/",
		domain,
		secure,
		h.config.Cookie.HttpOnly,
	)

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(h.config.Cookie.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no session found"})
		return
	}
	if err := h.authService.DeleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	domain, secure := h.resolveCookieSettings(c)
	c.SetCookie(
		h.config.Cookie.Name,
		"",
		-1,
		"/",
		domain,
		secure,
		h.config.Cookie.HttpOnly,
	)

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

func (h *AuthHandler) Validate(c *gin.Context) {
	sessionID, err := c.Cookie(h.config.Cookie.Name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no session found"})
		return
	}
	session, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": session.UserID})
}

func (h *AuthHandler) resolveCookieSettings(c *gin.Context) (domain string, secure bool) {
	isLocal := strings.Contains(c.Request.Host, "localhost") || strings.HasPrefix(c.Request.RemoteAddr, "127.")
	if isLocal {
		return "", false
	}
	return h.config.Cookie.Domain, h.config.Cookie.Secure
}

// func (h *AuthHandler) TestConnection(c *gin.Context) {
// 	users, err := h.authService.TestUserConnection(c.Request.Context())
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"message":     "connection successful",
// 		"users_count": len(users),
// 		"users":       users,
// 	})
// }
