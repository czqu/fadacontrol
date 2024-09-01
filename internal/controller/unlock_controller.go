package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/unlock"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UnlockController struct {
	u *unlock.UnLockService
}

func NewUnlockController(u *unlock.UnLockService) *UnlockController {
	return &UnlockController{u}
}

// @Summary	Unlock your computer
// @Produce	json
// @Param UserInfo body schema.PcUserInfo true "User Information"
// @Success	200		{object}	schema.ResponseData		"success"
// @Failure	400		{object}	schema.ResponseData			"The request is incorrect"
// @Failure	500		{object}	schema.ResponseData			"Internal errors"
// @Router		/unlock [post]
func (o *UnlockController) Unlock(c *gin.Context) {
	reqData := schema.PcUserInfo{}
	if err := c.Bind(&reqData); err != nil {
		c.Error(exception.ErrParameterError)
		return
	}
	if reqData.UserName == "" || reqData.Password == "" {
		c.Error(exception.ErrUsernamePasswordEmpty)
		return
	}
	if len(reqData.UserName) > 256 || len(reqData.Password) > 256 {
		c.Error(exception.ErrParameterLengthExceeds)
		return
	}

	logger.Info("Username and password information have been received")
	e := o.u.UnlockPc(reqData.UserName, reqData.Password)
	if e.NotEqual(exception.ErrSuccess) {
		logger.Info(e.Msg)
		c.Error(e)
		return
	}
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
	})

}
