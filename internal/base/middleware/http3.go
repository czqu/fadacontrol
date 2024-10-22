package middleware

import (
	"fadacontrol/internal/base/conf"
	"github.com/gin-gonic/gin"
	"strconv"
)

func UserHttp3() gin.HandlerFunc {

	return func(c *gin.Context) {
		if conf.Http3Enabled {
			c.Header("ALT-SVC", `h3=":"`+strconv.Itoa(conf.Http3Port))
			c.Next()
		}

	}
}
