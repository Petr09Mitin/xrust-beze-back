package httpparser

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit  int64 = 10
	defaultOffset int64 = 0
)

func GetLimitAndOffset(c *gin.Context) (limit int64, offset int64) {
	limit, err := strconv.ParseInt(strings.TrimSpace(c.Query("limit")), 10, 64)
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}
	offset, err = strconv.ParseInt(strings.TrimSpace(c.Query("offset")), 10, 64)
	if err != nil || offset < 0 {
		offset = defaultOffset
	}

	return limit, offset
}
