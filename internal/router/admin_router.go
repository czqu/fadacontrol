package router

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/middleware"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminRouter struct {
	swagHandler gin.HandlerFunc
	router      *gin.Engine
	u           *controller.UnlockController
	o           *controller.ControlPCController
	rc          *controller.RemoteController
	di          *controller.DiscoverController
}

func NewAdminRouter(rc *controller.RemoteController, u *controller.UnlockController, o *controller.ControlPCController, di *controller.DiscoverController) *AdminRouter {
	return &AdminRouter{router: gin.Default(), u: u, o: o, rc: rc, di: di}
}
func (d *AdminRouter) Register() {

	r := gin.Default()
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHandler())
	r.HandleMethodNotAllowed = true
	r.Delims("[[[", "]]]")
	r.NoRoute(d.get404Page)
	if adminSwagHandler != nil {
		r.GET("/swagger/*any", adminSwagHandler)
	}
	apiv1 := r.Group("/admin/api/v1")
	{
		apiv1.GET("/ping", controller.Ping)
		apiv1.POST("/control-pc/:action", d.o.ControlPC)
		apiv1.POST("/unlock", d.u.Unlock)
		apiv1.GET("/interface/:ip", d.o.GetInterfaceByIP)
		apiv1.GET("/interface/:ip/all", d.o.GetInterfaceByIPAll)
		apiv1.GET("/interface/", d.o.GetInterface)
		apiv1.GET("/remote-config", d.rc.GetRemoteConfig)
		apiv1.POST("/remote-config", d.rc.SetRemoteConfig)
		apiv1.GET("/remote-config/delay", d.rc.TestServerDelay)
		apiv1.POST("/discovery", d.di.SetDiscoverService)
		apiv1.GET("/discovery", d.di.GetDiscoverServiceConfig)
		//	apiv1.GET("/internal-cmd/", d.internal.GetInternalCommandEvents)

	}

	d.router = r
}
func (d *AdminRouter) GetRouter() *gin.Engine {
	return d.router
}
func (d *AdminRouter) get404Page(c *gin.Context) {
	c.JSON(http.StatusNotFound, schema.ResponseData{
		Code: exception.ErrResourceNotFound.Code,
		Msg:  exception.ErrResourceNotFound.Msg,
	})
}
