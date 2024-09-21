package admin_controller

import (
	"fadacontrol/internal/controller"
	"fadacontrol/internal/service/discovery_service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DiscoverController struct {
	di *discovery_service.DiscoverService
}

func NewDiscoverController(di *discovery_service.DiscoverService) *DiscoverController {
	return &DiscoverController{di: di}
}

// @Summary Get Discover Service Config
// @Description Get the Discovery Service configuration
// @Tags Discover
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} schema.ResponseData "success"
// @Failure 500 {object} schema.ResponseData "Server internal error"
// @Router /discovery/config [get]
func (d *DiscoverController) GetDiscoverServiceConfig(c *gin.Context) {

	ret, err := d.di.GetDiscoverConfig()
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, ret))
}

// @Summary Update Discover Service Configuration
// @Description Update the configuration of the Discover service with the provided settings.
// @Tags Discover
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param config body schema.DiscoverSchema true "New configuration settings"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /discovery/config [patch]
func (d *DiscoverController) PatchDiscoverServiceConfig(c *gin.Context) {

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(err)
		return
	}

	err := d.di.PatchDiscoverServiceConfig(data)
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))

}
