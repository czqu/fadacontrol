package middleware

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/schema"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, e := range c.Errors {
			err := e.Err
			if e, ok := err.(*exception.Exception); ok {
				status := http.StatusInternalServerError //The default is 500

				if e.Code == exception.ErrResourceNotFound.Code {
					status = http.StatusNotFound
				} else if (e.Code >= exception.UserErrorStart) && (e.Code <= exception.UserErrorEnd) {

					status = http.StatusBadRequest
				} else if e.Code == exception.ErrSuccess.Code {
					status = http.StatusOK
				} else {

				}

				c.JSON(status, schema.ResponseData{
					Code: e.Code,
					Msg:  e.Msg,
				})

			} else {
				c.JSON(http.StatusInternalServerError, schema.ResponseData{
					Code: exception.ErrSystemUnknownException.Code,
					Msg:  exception.ErrSystemUnknownException.Msg,
				})
			}
			return
		}
	}
}
