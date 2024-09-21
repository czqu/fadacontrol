package middleware

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {

			if err := recover(); err != nil {
				logger.Errorf("panic recover: %v", err)
				c.JSON(http.StatusInternalServerError, controller.GetGinErrorWithData(c, exception.ErrUnknownException, fmt.Sprintf("%v", err)))
			}
		}()
		c.Next()
	}

}
