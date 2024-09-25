package middleware

import "C"
import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/service/auth_service"
	"fadacontrol/internal/service/jwt_service"
	"fadacontrol/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

type JwtMiddleware struct {
	jw   *jwt_service.JwtService
	auth *auth_service.AuthService
}

func NewJwtMiddleware(jw *jwt_service.JwtService, auth *auth_service.AuthService) *JwtMiddleware {
	return &JwtMiddleware{jw: jw, auth: auth}
}

func (j *JwtMiddleware) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.SetRequestId(c)
		if j.isIgnoredPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			if j.auth.CheckHttpPermission("", c.Request.URL.Path, j.RequestMethodToAuthAction(c.Request.Method)) {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, controller.GetGinError(c, exception.ErrUserUnauthorizedAccess))
			c.Abort()
			return
		}

		claims, err := j.jw.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, controller.GetGinError(c, exception.ErrUserUnauthorizedAccess))
			c.Abort()
			return
		}

		c.Set("username", claims.Username)

		c.Next()
	}
}
func (j *JwtMiddleware) RequestMethodToAuthAction(method string) auth_service.Action {
	switch method {
	case http.MethodGet:
		return auth_service.Read
	case http.MethodPost:
		return auth_service.Write
	case http.MethodPut:
		return auth_service.Write
	case http.MethodDelete:
		return auth_service.Write
	case http.MethodOptions:
		return auth_service.Read
	default:
		return auth_service.Write
	}

}
func (j *JwtMiddleware) isIgnoredPath(path string) bool {
	for _, pattern := range conf.IgnoredPaths {
		matched, _ := regexp.MatchString(pattern, path)
		if matched {
			return true
		}
	}
	return false
}
