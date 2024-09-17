package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/schema/remote_schema"
	"fadacontrol/internal/service/remote_service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type RemoteController struct {
	db  *gorm.DB
	rcs *remote_service.RemoteService
}

func NewRemoteController(db *gorm.DB, rcs *remote_service.RemoteService) *RemoteController {
	return &RemoteController{db: db, rcs: rcs}
}

func (o *RemoteController) SetRemoteConfig(c *gin.Context) {
	reqData := remote_schema.RemoteConfigReqDTO{}

	if err := c.Bind(&reqData); err != nil {
		c.Error(exception.ErrParameterError)
		return
	}
	err := o.rcs.UpdateData(reqData)
	if err != nil {
		c.Error(exception.ErrParameterError)
		return
	}

	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
		Data: "",
	})
	o.rcs.RestartService()
}
func (o *RemoteController) GetRemoteConfig(c *gin.Context) {
	remoteConfig, err := o.rcs.GetData()
	if err != nil {
		c.Error(exception.ErrParameterError)
		return
	}
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
		Data: remoteConfig,
	})
}
func (o *RemoteController) TestServerDelay(c *gin.Context) {
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
		Data: o.rcs.TestServerDelay() / 2,
	})
}
