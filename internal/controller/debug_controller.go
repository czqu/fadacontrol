package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary	Ping
// @Produce	json
// @Success	200		"success"
// @Router		/ping [get]
func Ping(c *gin.Context) {
	logger.Info("receive ping")
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
	})
}
