package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetRequestId(c *gin.Context) string {
	requestId, ok := c.Get("request_id")
	if !ok {
		requestId = ""
	}
	var r string
	switch requestId.(type) {
	case string:
		r = requestId.(string)
	default:
		r = ""
	}
	return r
}

func SetRequestId(c *gin.Context) {
	c.Set("request_id", uuid.New().String())
}
