package controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/discovery"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type DiscoverController struct {
	db     *gorm.DB
	config entity.DiscoverConfig
}

func NewDiscoverController(db *gorm.DB) *DiscoverController {
	return &DiscoverController{db: db}
}

func (d *DiscoverController) SetDiscoverService(c *gin.Context) {
	resp := &schema.DiscoverSchema{}
	err := c.Bind(resp)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if resp.Enabled {
		d.config.Enabled = true
		discovery.StartBroadcast()
	} else {
		d.config.Enabled = false
		discovery.StopBroadcast()
	}
	d.db.Save(&d.config)

	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
	})

}
func (d *DiscoverController) GetDiscoverServiceConfig(c *gin.Context) {
	if err := d.db.First(&d.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)

	}
	res := &schema.DiscoverSchema{
		Enabled: d.config.Enabled,
	}
	c.JSON(http.StatusOK, schema.ResponseData{
		Code: exception.ErrSuccess.Code,
		Msg:  exception.ErrSuccess.Msg,
		Data: res,
	})
}
