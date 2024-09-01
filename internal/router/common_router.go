package router

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/middleware"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"net/http"
)
import "github.com/gin-gonic/gin"

type CommonRouter struct {
	swagHandler gin.HandlerFunc
	router      *gin.Engine

	u  *controller.UnlockController
	o  *controller.ControlPCController
	cu *controller.CustomCommandController
}

func NewCommonRouter(cu *controller.CustomCommandController, u *controller.UnlockController, o *controller.ControlPCController) *CommonRouter {
	return &CommonRouter{router: gin.Default(), u: u, o: o, cu: cu}
}
func (d *CommonRouter) Register() {

	r := gin.Default()
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHandler())
	r.HandleMethodNotAllowed = true
	r.Delims("[[[", "]]]")
	r.NoRoute(d.get404Page)
	if commonSwagHandler != nil {
		r.GET("/swagger/*any", commonSwagHandler)
	}
	apiv1 := r.Group("/api/v1")

	{

		apiv1.GET("/ping", controller.Ping)
		apiv1.POST("/control-pc/:action", d.o.ControlPC)
		apiv1.POST("/unlock", d.u.Unlock)
		apiv1.GET("/interface/:ip", d.o.GetInterfaceByIP)
		apiv1.GET("/interface/:ip/all", d.o.GetInterfaceByIPAll)
		apiv1.GET("/interface/", d.o.GetInterface)
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
