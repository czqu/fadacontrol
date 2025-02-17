package admin_router

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/middleware"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/controller/admin_controller"
	"fadacontrol/internal/controller/common_controller"
	"fadacontrol/internal/schema"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminRouter struct {
	swagHandler gin.HandlerFunc
	router      *gin.Engine
	u           *common_controller.UnlockController
	o           *common_controller.ControlPCController
	rc          *admin_controller.RemoteController
	di          *admin_controller.DiscoverController
	auth        *common_controller.AuthController
	jwt         *middleware.JwtMiddleware
	_sys        *common_controller.SystemController
	_http       *admin_controller.HttpController
	_de         *common_controller.DebugController
}

func NewAdminRouter(_de *common_controller.DebugController, _http *admin_controller.HttpController, sys *common_controller.SystemController, jwt *middleware.JwtMiddleware, rc *admin_controller.RemoteController, u *common_controller.UnlockController, o *common_controller.ControlPCController, di *admin_controller.DiscoverController, auth *common_controller.AuthController) *AdminRouter {
	return &AdminRouter{router: gin.Default(), u: u, o: o, rc: rc, di: di, auth: auth, jwt: jwt, _sys: sys, _http: _http, _de: _de}
}

var swagHandler gin.HandlerFunc

func (d *AdminRouter) Register() {

	r := gin.Default()
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHandler())
	r.Use(d.jwt.JWTAuthMiddleware())
	r.HandleMethodNotAllowed = true
	r.Delims("[[[", "]]]")
	r.NoRoute(d.get404Page)
	if swagHandler != nil && (version.Edition != "release") {
		r.GET("/swagger/*any", swagHandler)
	}
	apiv1 := r.Group("/admin/api/v1")
	{
		apiv1.GET("/ping", d._de.Ping)
		apiv1.POST("/control-pc/:action", d.o.ControlPC)
		apiv1.POST("/unlock", d.u.Unlock)
		apiv1.GET("/interface/:ip", d.o.GetInterfaceByIP)
		apiv1.GET("/interface/:ip/all", d.o.GetInterfaceByIPAll)
		apiv1.GET("/interface/", d.o.GetInterface)
		apiv1.GET("/info", d._sys.GetSoftwareInfo)
		apiv1.GET("/info/check_update", d._sys.CheckUpdate)
		apiv1.GET("/info/language", d._sys.GetLanguage)
		apiv1.PATCH("/info/language", d._sys.SetLanguage)
		apiv1.GET("/logs", d._sys.GetLog)
		apiv1.GET("/logs/:module", d._sys.GetLog)
		apiv1.POST("/power-saving", d._sys.SetPowerSavingMode)
		apiv1.GET("/power-saving/status", d._sys.GetPowerSavingModeStatus)
		//	apiv1.GET("/internal-cmd/", d.internal.GetInternalCommandEvents)

		//admin
		apiv1.GET("/discovery/config", d.di.GetDiscoverServiceConfig)
		apiv1.PATCH("/discovery/config", d.di.PatchDiscoverServiceConfig)
		apiv1.POST("/discovery/restart", d.di.RestartDiscoverService)

		apiv1.GET("/remote/config", d.rc.GetRemoteConnectConfig)
		apiv1.PATCH("/remote/config", d.rc.PatchRemoteConnectConfig)
		apiv1.PUT("/remote/config", d.rc.UpdateRemoteConnectConfig)
		apiv1.POST("/remote/restart", d.rc.RestartRemoteService)

		apiv1.GET("/http/config", d._http.GetHttpConfig)
		apiv1.PATCH("/http/config", d._http.PatchHttpConfig)
		apiv1.PUT("/http/config", d._http.UpdateHttpConfig)

		apiv1.POST("/sys/stop", d._http.StopService)

		apiv1.POST("/login", d.auth.Login)
	}

	d.router = r
}
func (d *AdminRouter) GetRouter() *gin.Engine {
	return d.router
}
func (d *AdminRouter) get404Page(c *gin.Context) {
	c.JSON(http.StatusNotFound, schema.ResponseData{
		Code: exception.ErrUserResourceNotFound.Code,
		Msg:  exception.ErrUserResourceNotFound.Msg,
	})
}
func SetSwagHandler(handler gin.HandlerFunc) {
	swagHandler = handler
}
