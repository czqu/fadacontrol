package router

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/middleware"
	"fadacontrol/internal/controller/common_controller"
	"fadacontrol/internal/schema"
	"net/http"
)
import "github.com/gin-gonic/gin"

type CommonRouter struct {
	swagHandler gin.HandlerFunc
	router      *gin.Engine

	u    *common_controller.UnlockController
	o    *common_controller.ControlPCController
	cu   *common_controller.CustomCommandController
	auth *common_controller.AuthController
	jwt  *middleware.JwtMiddleware
	sys  *common_controller.SysInfoController
}

func NewCommonRouter(sys *common_controller.SysInfoController, jwt *middleware.JwtMiddleware, auth *common_controller.AuthController, cu *common_controller.CustomCommandController, u *common_controller.UnlockController, o *common_controller.ControlPCController) *CommonRouter {
	return &CommonRouter{router: gin.Default(), u: u, o: o, cu: cu, auth: auth, jwt: jwt, sys: sys}
}
func (d *CommonRouter) Register() {

	r := gin.Default()
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHandler())
	r.Use(d.jwt.JWTAuthMiddleware())
	r.HandleMethodNotAllowed = true
	r.Delims("[[[", "]]]")
	r.NoRoute(d.get404Page)
	if commonSwagHandler != nil {
		r.GET("/swagger/*any", commonSwagHandler)
	}
	apiv1 := r.Group("/api/v1")

	{

		apiv1.GET("/ping", common_controller.Ping)
		apiv1.POST("/control-pc/:action", d.o.ControlPC)
		apiv1.POST("/unlock", d.u.Unlock)
		apiv1.GET("/interface/:ip", d.o.GetInterfaceByIP)
		apiv1.GET("/interface/:ip/all", d.o.GetInterfaceByIPAll)
		apiv1.GET("/interface/", d.o.GetInterface)
		apiv1.POST("/login", d.auth.Login)
		apiv1.GET("/info", d.sys.GetSoftwareInfo)
		//apiv1.POST("/execute", d.cu.Execute)
		//apiv1.GET("/execute/:id", d.cu.ExecResult)

	}

	d.router = r
}
func (d *CommonRouter) GetRouter() *gin.Engine {
	return d.router
}
func (d *CommonRouter) get404Page(c *gin.Context) {
	c.JSON(http.StatusNotFound, schema.ResponseData{
		Code: exception.ErrResourceNotFound.Code,
		Msg:  exception.ErrResourceNotFound.Msg,
	})
}
