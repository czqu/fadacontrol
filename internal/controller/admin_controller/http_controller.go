package admin_controller

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema/http_schema"
	"fadacontrol/internal/service/http_service"
	"fadacontrol/pkg/goroutine"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type HttpController struct {
	_db         *gorm.DB
	hs          *http_service.HttpService
	_exitSignal *conf.ExitChanStruct
}

func NewHttpController(_exitSignal *conf.ExitChanStruct, db *gorm.DB, hs *http_service.HttpService) *HttpController {
	return &HttpController{_exitSignal: _exitSignal, _db: db, hs: hs}
}

// @Summary Get HTTP Configuration
// @Description Retrieve the current HTTP configuration based on the provided type.
// @Tags HTTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param type query string true "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)"
// @Success 200 {object} schema.ResponseData "Successfully retrieved configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /http/config [get]
func (h *HttpController) GetHttpConfig(c *gin.Context) {
	query := c.Query("type")
	resp, err := h.hs.GetHttpConfig(query)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, resp))
}

// @Summary Update HTTP Configuration
// @Description Update the HTTP configuration based on the provided type and settings.
// @Tags HTTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param type query string true "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)"
// @Param config body http_schema.HttpConfigRequest true "Configuration settings"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /http/config [put]
func (h *HttpController) UpdateHttpConfig(c *gin.Context) {
	query := c.Query("type")
	if query == http_service.HttpsServiceApi {
		var request http_schema.HttpsConfigRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(err)
			return
		}
		err := h.hs.UpdateHttpConfig(&request, query)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return
	}
	if query == http_service.HttpServiceApi {
		var request http_schema.HttpConfigRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(err)
			return
		}
		err := h.hs.UpdateHttpConfig(&request, query)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return
	}

	c.Error(exception.ErrUserParameterError)

}

// @Summary Patch HTTP Configuration
// @Description Partially update the HTTP configuration based on the provided type and settings.
// @Tags HTTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param type query string true "Configuration type (HTTP_SERVICE_API or HTTPS_SERVICE_API)"
// @Param config body http_schema.HttpConfigRequest true "Partial configuration settings"
// @Success 200 {object} schema.ResponseData "Successfully updated configuration."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /http/config [patch]
func (h *HttpController) PatchHttpConfig(c *gin.Context) {
	query := c.Query("type")
	if query == http_service.HttpsServiceApi {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Error(err)
			return
		}

		err := h.hs.PatchHttpConfig(data, query)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return

	}
	if query == http_service.HttpServiceApi {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Error(err)
			return
		}
		err := h.hs.PatchHttpConfig(data, query)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return

	}

	c.Error(exception.ErrUserParameterError)
}

// @Summary Exit Service
// @Description Exit the server
// @Tags HTTP
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData "Service restarted successfully."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /sys/stop [post]
func (h *HttpController) StopService(c *gin.Context) {
	goroutine.RecoverGO(func() {
		time.Sleep(1 * time.Second)
		h._exitSignal.ExitChan <- 0
	})

	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}
