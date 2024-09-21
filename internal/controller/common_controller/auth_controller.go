package common_controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/jwt_service"
	"fadacontrol/internal/service/user_service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthController struct {
	s   *user_service.UserService
	jwt *jwt_service.JwtService
}

func NewAuthController(s *user_service.UserService, jwt *jwt_service.JwtService) *AuthController {
	return &AuthController{s: s, jwt: jwt}
}

// @Summary User Login
// @Description Authenticate user with username and password to obtain a JWT token.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param loginData body schema.LoginRequest true "User login credentials"
// @Success 200 {object} schema.ResponseData "Successfully authenticated."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /login [post]
func (u *AuthController) Login(c *gin.Context) {

	loginData := &schema.LoginRequest{}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.Error(exception.ErrParameterError)
		return
	}

	user, err := u.s.Login(loginData.Username, loginData.Password)
	if err != nil {
		c.Error(exception.ErrLogonFailure)
		return
	}
	token, err := u.jwt.GenerateToken(user.Username)
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, schema.TokenResponse{Token: token}))
}
