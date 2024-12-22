package common_controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/unlock"
	"fmt"
	"github.com/getsentry/sentry-go"
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
	ctx := c.Request.Context()
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Check the concurrency guide for more details: https://docs.sentry.io/platforms/go/concurrency/
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	options := []sentry.SpanOption{
		// Set the OP based on values from https://develop.sentry.dev/sdk/performance/span-operations/
		sentry.WithOpName("http.server"),
		sentry.ContinueFromRequest(c.Request),
		sentry.WithTransactionSource(sentry.SourceURL),
	}
	transaction := sentry.StartTransaction(ctx,
		fmt.Sprintf("HTTP: %s %s", c.Request.Method, c.Request.URL.Path),
		options...,
	)
	defer transaction.Finish()

	reqData := schema.PcUserInfo{}
	if err := c.Bind(&reqData); err != nil {
		c.Error(exception.ErrUserParameterError)
		return
	}
	if reqData.UserName == "" || reqData.Password == "" {
		c.Error(exception.ErrUsernamePasswordEmpty)
		return
	}
	if len(reqData.UserName) > 256 || len(reqData.Password) > 256 {
		c.Error(exception.ErrUserParameterLengthExceeds)
		return
	}

	logger.Info("Username and password information have been received")
	e := o.u.UnlockPc(reqData.UserName, reqData.Password)
	if exception.ErrSuccess.NotEqual(e) {
		logger.Info(e.Msg)
		c.Error(e)
		return
	}
	logger.Info("Unlock successful")
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))

}
