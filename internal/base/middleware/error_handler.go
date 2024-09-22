package middleware

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() //must be called first
		if c.Writer.Status() == http.StatusMethodNotAllowed {

			c.JSON(http.StatusMethodNotAllowed, controller.GetGinError(c, exception.ErrUserMethodNotAllowed))
			return
		}
		if len(c.Errors) == 0 {

			return
		}
		code := exception.ErrUnknownException.Code
		msg := exception.ErrUnknownException.Msg
		for _, e := range c.Errors {
			err := e.Err
			switch ex := err.(type) {
			case *exception.Exception:
				code = ex.Code
				msg = ex.Msg
			case error:
				msg = ex.Error()
			default:
				msg = ex.Error()
			}

			status := http.StatusInternalServerError //The default is 500
			if code == exception.ErrUserResourceNotFound.Code {
				status = http.StatusNotFound
			} else if (code >= exception.UserErrorStart) && (code <= exception.UserErrorEnd) {

				status = http.StatusBadRequest
			} else if code == exception.ErrSuccess.Code {
				status = http.StatusOK
			} else {

			}

			c.JSON(status, controller.GetGinError(c, &exception.Exception{Code: code, Msg: msg}))

			return
		}
	}
}
