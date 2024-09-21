package common_controller

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary	Ping
// @Produce	json
// @Success	200		"success"
// @Router		/ping [get]
func Ping(c *gin.Context) {
	logger.Debug("receive ping")
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}
