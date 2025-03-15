package common

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HTTPContextKey string

var (
	JWTAuthHeader                  = "Authorization"
	JaegerHeader                   = "Uber-Trace-Id"
	ChannelIdHeader                = "X-Channel-Id"
	ChannelKey      HTTPContextKey = "channel_key"
	UserKey         HTTPContextKey = "user_key"
)

func MaxAllowed(n int64) gin.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }
	return func(c *gin.Context) {
		acquire()       // before request
		defer release() // after request
		c.Next()

	}
}

func CorsMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", JWTAuthHeader},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

func LoggingMiddleware(logger HttpLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		duration := getDurationInMillseconds(start)

		logger.Info("",
			slog.Float64("duration_ms", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.RequestURI),
			slog.Int("status", c.Writer.Status()),
			slog.String("referrer", c.Request.Referer()),
			slog.String("trace_id", getTraceID(c)))
	}
}

func getTraceID(c *gin.Context) string {
	identifier := c.Request.Header.Get(JaegerHeader)
	vals := strings.Split(identifier, ":")
	if len(vals) == 4 {
		return vals[0]
	}
	return ""
}

func getDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}

func extractTokenFromHeader(r *http.Request) string {
	bearToken := r.Header.Get(JWTAuthHeader)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
