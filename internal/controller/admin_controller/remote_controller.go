package admin_controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/schema/remote_schema"
	"fadacontrol/internal/service/remote_service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
func (o *RemoteController) GetRemoteConfig(c *gin.Context) {

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
func (o *RemoteController) UpdateRemoteConfig(c *gin.Context) {
	var request remote_schema.RemoteConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(exception.ErrUserParameterError)
		return
	}
	err := o.rcs.UpdateRemoteConfig(&request)
	if err != nil {
		c.Error(err)
		return
	}
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
func (o *RemoteController) PatchRemoteConfig(c *gin.Context) {

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(err)
		return
	}

	err := o.rcs.PatchRemoteConfig(data)
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}
func (o *RemoteController) GetRemoteApiServerConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Error(err)
		return
	}
	remoteResp, err := o.rcs.GetRemoteApiServerConfig(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, remoteResp))
}
func (o *RemoteController) UpdateRemoteApiServerConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Error(err)
		return
	}
	var request remote_schema.RemoteApiServerConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(err)
		return
	}
	ret, err := o.rcs.UpdateRemoteApiServerConfig(id, &request)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, ret))

}

func (o *RemoteController) GetCredential(c *gin.Context) {
	credential, err := o.rcs.GetCredential()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, credential))
}
func (o *RemoteController) RefreshCredential(c *gin.Context) {
	credential, err := o.rcs.RefreshCredential()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, credential))
}

// @Summary Restart Remote Service
// @Description Restart the remote service.
// @Tags Remote
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData "Service restarted successfully."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /remote/restart [post]
func (o *RemoteController) RestartRemoteService(c *gin.Context) {

	err := o.rcs.RestartService()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}

func (o *RemoteController) GetNowServerDelay(c *gin.Context) {
	delay, err := o.rcs.TestServerDelay()
	if err != nil {
		c.Error(err)
		return
	}
	logger.Debug("delay:", delay)
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
		Data: delay / 2,
	})
}

//func (o *RemoteController) SetRemoteConfig(c *gin.Context) {
//	reqData := remote_schema.RemoteConfigReqDTO{}
//
//	if err := c.Bind(&reqData); err != nil {
//		c.Error(exception.ErrUserParameterError)
//		return
//	}
//	err := o.rcs.UpdateData(reqData)
//	if err != nil {
//		c.Error(exception.ErrUserParameterError)
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
