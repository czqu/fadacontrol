package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/pkg/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"net/http"
)

type ControlPCController struct {
	p *control_pc.ControlPCService
}

func NewControlPCController(p *control_pc.ControlPCService) *ControlPCController {
	return &ControlPCController{p: p}
}

// ControlPC Control computer interface
//
//	@Summary		Control  computer
//	@Description	Control the operation of the computer according to the transmitted parameters
//	@Accept			json
//	@Produce		json
//	@Param			action	path		string					true	"The type of operation（shutdown、standby、lock）"	Enums(shutdown, standby, lock)
//	@Success		200		{object}	schema.ResponseData	"success"
//	@Failure		400		{object}	schema.ResponseData	"Invalid action type"
//	@Failure		500		{object}	schema.ResponseData		"The operation failed"
//	@Router			/control-pc/{action}/ [post]
func (o *ControlPCController) ControlPC(c *gin.Context) {
	action := c.Param("action")

	var result string
	var ret *exception.Exception
	switch action {
	case "shutdown":

		ret = o.p.Shutdown()
		result = "Shutdown"
	case "standby":
		ret = o.p.Standby()
		result = "Standby"
	case "lock":
		ret = o.p.LockWindows(true)
		result = "Lock"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code": exception.ErrParameterError,
			"msg":  "Invalid action specified",
		})
		return
	}

	if ret != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": exception.UnknownError,
			"msg":  fmt.Sprintf("Failed to %s, error: %s", result, exception.ErrUnauthorizedAccess),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  fmt.Sprintf("%s command sent successfully", result),
	})
}

// GetInterface Returns a valid network interface
// @Summary Returns a valid network interface
// @Description Based on the specified IP version type, a list of valid network interfaces is returned
// @Tags Network interfaces
// @Accept json
// @Produce json
// @Param type query string false "IP Version Type (4 or 6)" default(4) Enums(4, 6)
// @Success 200 {object} schema.ResponseData "A list of valid network interfaces is successfully returned"
// @Failure 400 {object} schema.ResponseData "The request parameter is incorrect"
// @Failure 500 {object} schema.ResponseData "Server internal error"
// @Router /control-pc/interface [get]
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
	c.JSON(http.StatusOK,
		schema.ResponseData{
			Code: exception.ErrSuccess.Code,
			Msg:  exception.ErrSuccess.Msg,
			Data: ifces,
		})
}

// @Summary	Obtain the MAC address based on the IP address
// @Produce	json
// @Success	200	{object}	schema.ResponseData			"success"
// @Failure	400	{object}	schema.ResponseData				"The request is incorrect"
// @Failure	500	{object}	schema.ResponseData						"Internal errors"
// @Router		/interface/{ip} [get]
func (o *ControlPCController) GetInterfaceByIP(c *gin.Context) {
	o.getInterfaceOld(c, false)

}

// @Summary	Obtain the interface information based on the IP address
// @Produce	json
// @Success	200								"success"
// @Failure	400				"The request is incorrect"
// @Failure	500							"Internal errors"
// @Router		/interface/{ip}/all [get]
func (o *ControlPCController) GetInterfaceByIPAll(c *gin.Context) {
	o.getInterfaceOld(c, true)
}

func (o *ControlPCController) getInterfaceOld(c *gin.Context, all bool) {
	ip := c.Param("ip")
	ifces, err := utils.GetValidInterface(utils.UNSET)
	if err != nil {
		c.Error(exception.ErrUnknownException)
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
		c.Error(exception.ErrResourceNotFound)
		return
	}

	if all {
		c.JSON(http.StatusOK, schema.ResponseData{
			Code: exception.ErrSuccess.Code,
			Msg:  exception.ErrSuccess.Msg,
			Data: info,
		})
	} else {

		c.JSON(http.StatusOK,
			schema.ResponseData{
				Code: exception.ErrSuccess.Code,
				Msg:  exception.ErrSuccess.Msg,
				Data: info.MACAddr,
			})

	}
}
