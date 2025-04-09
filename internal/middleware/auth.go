package middleware

import (
	"net/http"

	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("skill_sharing_token") // хардкод, придумать, как избавиться
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
