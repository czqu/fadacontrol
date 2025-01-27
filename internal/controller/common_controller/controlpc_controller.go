package common_controller

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
)

type ControlPCController struct {
	p   *control_pc.ControlPCService
	ctx context.Context
}

func NewControlPCController(ctx context.Context, p *control_pc.ControlPCService) *ControlPCController {
	return &ControlPCController{ctx: ctx, p: p}
}

// ControlPC Control computer interface
//
//			@Summary		Control  computer
//			@Description	Control the operation of the computer according to the transmitted parameters
//			@Accept			json
//			@Produce		json
//			@Param			action	path		string					true	"The type of operation（shutdown、standby、lock）"	Enums(shutdown, standby, lock)
//	     @Param			delay		query		string	false	"Delay time in seconds,only valid when the action is shutdown、standby"
//		    @Param			shutdown_type	query		string	false	"The type of shutdown"	sys.ShutdownType
//			@Success		200		{object}	schema.ResponseData	"success"
//			@Failure		400		{object}	schema.ResponseData	"Invalid action type"
//			@Failure		500		{object}	schema.ResponseData		"The operation failed"
//			@Router			/control-pc/{action}/ [post]
//
// @Security ApiKeyAuth
func (o *ControlPCController) ControlPC(c *gin.Context) {
	delayS := c.Query("delay")
	action := c.Param("action")

	var delaySec int
	delaySec, err := strconv.Atoi(delayS)
	if err != nil {
		delaySec = 5
	}
	var ret error
	switch action {
	case "shutdown":

		tpe := c.DefaultQuery("shutdown_type", strconv.Itoa(int(sys.S_E_FORCE_SHUTDOWN)))
		shutdownType, err := strconv.Atoi(tpe)
		if err != nil {
			c.Error(exception.ErrUserParameterError)
			return
		}
		goroutine.RecoverGO(func() {
			time.Sleep(time.Duration(delaySec) * time.Second)
			ret = o.p.Shutdown(sys.ShutdownType(shutdownType))
		})

	case "standby":
		goroutine.RecoverGO(func() {
			time.Sleep(time.Duration(delaySec) * time.Second)
			ret = o.p.Standby()
		})

	case "lock":
		_conf := utils.GetValueFromContext(o.ctx, constants.ConfKey, conf.NewDefaultConf())
		if _conf.StartMode == conf.ServiceMode {
			ret = o.p.LockWindows(true)
		} else {
			ret = o.p.LockWindows(false)
		}

	default:
		c.Error(exception.ErrUserParameterError)
		return
	}

	if ret != nil && exception.ErrSuccess.NotEqual(ret) {
		c.Error(ret)
		return
	}

	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}

// GetInterface Returns a valid network interface
// @Summary Returns a valid network interface
// @Description Based on the specified IP version type, a list of valid network interfaces is returned
// @Tags Network interfaces
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param type query string false "IP Version Type (4 or 6)" default(4) Enums(4, 6)
// @Success 200 {object} schema.ResponseData "A list of valid network interfaces is successfully returned"
// @Failure 400 {object} schema.ResponseData "The request parameter is incorrect"
// @Failure 500 {object} schema.ResponseData "Server internal error"
// @Router /interface [get]
func (o *ControlPCController) GetInterface(c *gin.Context) {
	typeParam := c.DefaultQuery("type", "4")
	t := utils.IPV4
	if typeParam == "6" {
		t = utils.IPV6

	}

	ifces, err := utils.GetValidInterface(t)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, &ifces))

}

// @Summary	Obtain the MAC address based on the IP address
// @Produce	json
// @Security ApiKeyAuth
// @Success	200	{object}	schema.ResponseData			"success"
// @Failure	400	{object}	schema.ResponseData				"The request is incorrect"
// @Failure	500	{object}	schema.ResponseData						"Internal errors"
// @Param		ip								path		string	true	"IP address"
// @Router		/interface/{ip} [get]
func (o *ControlPCController) GetInterfaceByIP(c *gin.Context) {
	o.getInterfaceOld(c, false)

}

// @Summary	Obtain the interface information based on the IP address
// @Produce	json
// @Security ApiKeyAuth
// @Success	200								"success"
// @Failure	400				"The request is incorrect"
// @Failure	500							"Internal errors"
// @Param		ip								path		string	true	"IP address"
// @Router		/interface/{ip}/all [get]
func (o *ControlPCController) GetInterfaceByIPAll(c *gin.Context) {
	o.getInterfaceOld(c, true)
}

func (o *ControlPCController) getInterfaceOld(c *gin.Context, all bool) {
	ip := c.Param("ip")
	ifces, err := utils.GetValidInterface(utils.UNSET)
	if err != nil {
		c.Error(err)
		return
	}
	ipMacMap := make(map[string]utils.Interface)
	for _, v := range ifces {
		ips := v.IPAddresses
		for _, e := range ips {
			ipMacMap[e.String()] = v
		}

	}
	info, ok := ipMacMap[ip]

	if !ok || info.MACAddr == "" {
		c.Error(exception.ErrUserResourceNotFound)
		return
	}

	if all {
		c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, &info))
	} else {

		c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, &info.MACAddr))

	}
}
