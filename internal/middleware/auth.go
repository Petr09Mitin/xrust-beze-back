package middleware

import (
	"net/http"

	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
)

const (
	SkillSharingTokenCookieKey = "skill_sharing_token"
)

func AuthMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SkillSharingTokenCookieKey)
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		resp, err := authClient.ValidateSession(
			c.Request.Context(),
			&authpb.SessionRequest{SessionId: sessionID},
		)
		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		c.Set("user_id", resp.UserId)
		c.Next()
	}
}

func CheckAuthMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SkillSharingTokenCookieKey)
		if err != nil || sessionID == "" {
			c.Next()
			return
		}

		resp, err := authClient.ValidateSession(
			c.Request.Context(),
			&authpb.SessionRequest{SessionId: sessionID},
		)
		if err == nil && resp.Valid {
			c.Set("user_id", resp.UserId)
		}
		// если сессия невалидна, то не устанавливаем user_id
		c.Next()
	}
}

func GetUserIDFromGinContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}
	return userIDStr, true
}
