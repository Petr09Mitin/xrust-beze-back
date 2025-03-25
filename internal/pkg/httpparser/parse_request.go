package httpparser

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func GetLimitAndOffset(c *gin.Context) (limit int64, offset int64) {
	limit, err := strconv.ParseInt(strings.TrimSpace(c.Query("limit")), 10, 64)
	if err != nil {
		limit = 0
	}
	offset, err = strconv.ParseInt(strings.TrimSpace(c.Query("offset")), 10, 64)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
