package admin_controller

import (
	"fadacontrol/internal/controller"
	"fadacontrol/internal/service/bluetooth_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BluetoothController struct {
	bt *bluetooth_service.BluetoothService
}

func NewBluetoothController(bt *bluetooth_service.BluetoothService) *BluetoothController {
	return &BluetoothController{bt: bt}
}

// @Summary GetBluetoothConfig
// @Description GetBluetoothConfig
// @Tags Bluetooth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.BluetoothSchema "Success"
// @Failure 500 {object} controller.ErrorResponse "Internal Server Error"
// @Router /bluetooth/config [get]
func (b *BluetoothController) GetBluetoothConfig(c *gin.Context) {
	ret, err := b.bt.GetBluetoothConfig()
	if err != nil {
		return
	}
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, ret))
}

// @Summary PatchBluetoothConfig
// @Description PatchBluetoothConfig
// @Tags Bluetooth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.BluetoothSchema true "Bluetooth Config"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /bluetooth/config [patch]
func (b *BluetoothController) PatchBluetoothConfig(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(err)
		return
	}
	if err := b.bt.PatchBluetoothConfig(&data); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, nil))
}

// @Summary RestartBluetoothService
// @Description RestartBluetoothService
// @Tags Bluetooth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /bluetooth/restart [post]
func (b *BluetoothController) RestartBluetoothService(c *gin.Context) {
	if err := b.bt.RestartBoothService(); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, nil))
}
