package auth_http

import (
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	"github.com/rs/zerolog"
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
	logger      zerolog.Logger
}

func NewAuthHandler(router *gin.Engine, authService auth.AuthService, config *config.Auth, logger zerolog.Logger) {
	handler := &AuthHandler{
		authService: authService,
		config:      config,
		logger:      logger,
	}

	router.Use(middleware.CORSMiddleware())
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
		h.logger.Info().Err(err).Msg("unable to bind json body in registration")
		custom_errors.WriteHTTPError(c, custom_errors.ErrInvalidBody)
		return
	}
	if err := req.Validate(); err != nil {
		if validationResp := validation.BuildValidationError(err); validationResp != nil {
			c.JSON(http.StatusBadRequest, validationResp)
			return
		}
		custom_errors.WriteHTTPError(c, err)
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
		custom_errors.WriteHTTPError(c, err)
		return
	}

	session, err := h.authService.CreateSession(c.Request.Context(), user.ID.Hex())
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
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
		h.logger.Info().Err(err).Msg("unable to bind json body in login")
		custom_errors.WriteHTTPError(c, custom_errors.ErrInvalidBody)
		return
	}

	session, user, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
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
		h.logger.Info().Err(err).Msg("unable to get cookie in logout")
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoAuthCookie)
		return
	}
	if err := h.authService.DeleteSession(c.Request.Context(), sessionID); err != nil {
		custom_errors.WriteHTTPError(c, err)
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
		h.logger.Info().Err(err).Msg("unable to get cookie in validate")
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoAuthCookie)
		return
	}
	session, user, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error().Err(err).Msg("invalid session in validate")
		custom_errors.WriteHTTPError(c, custom_errors.ErrInvalidAuthCookie)
		return
	}
	if session == nil {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoAuthCookie)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": session.UserID, "user": user})
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
