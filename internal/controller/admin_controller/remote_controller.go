package admin_controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
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

// @Summary Get Remote Connect Configuration
// @Description Retrieve the current configuration for remote connections.
// @Tags Remote
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData	 "Successfully retrieved configuration."
// @Failure 400 {object} schema.ResponseData	 "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData	"Internal Server Error"
// @Router /remote/config [get]
func (o *RemoteController) GetRemoteConnectConfig(c *gin.Context) {

	remoteResp, err := o.rcs.GetConfig()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, remoteResp))
}

// @Summary Update Remote Connect Configuration
// @Description Update the configuration for remote connections with the provided settings.
// @Tags Remote
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param config body remote_schema.RemoteConnectConfigRequest true "New configuration settings"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /remote/config [put]
func (o *RemoteController) UpdateRemoteConnectConfig(c *gin.Context) {
	var request remote_schema.RemoteConnectConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(exception.ErrParameterError)
		return
	}
	o.rcs.UpdateRemoteConnectConfig(&request)
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}

// @Summary Patch Remote Connect Configuration
// @Description Partially update the configuration for remote connections with the provided settings.
// @Tags Remote
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param config body remote_schema.RemoteConnectConfigRequest true "Partial configuration settings"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /remote/config [patch]
func (o *RemoteController) PatchRemoteConnectConfig(c *gin.Context) {

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(err)
		return
	}

	err := o.rcs.PatchRemoteConnectConfig(data)
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}

//	func (o *RemoteController) TestServerDelay(c *gin.Context) {
//		c.JSON(http.StatusOK, schema.ResponseData{
//			Code: exception.ErrSuccess.Code,
//			Msg:  exception.ErrSuccess.Msg,
//			Data: o.rcs.TestServerDelay() / 2,
//		})
//	}
//func (o *RemoteController) SetRemoteConfig(c *gin.Context) {
//	reqData := remote_schema.RemoteConfigReqDTO{}
//
//	if err := c.Bind(&reqData); err != nil {
//		c.Error(exception.ErrParameterError)
//		return
//	}
//	err := o.rcs.UpdateData(reqData)
//	if err != nil {
//		c.Error(exception.ErrParameterError)
//		return
//	}
//
//	c.JSON(http.StatusOK, schema.ResponseData{
//		Code: exception.ErrSuccess.Code,
//		Msg:  exception.ErrSuccess.Msg,
//		Data: "",
//	})
//	o.rcs.RestartService()
//}
