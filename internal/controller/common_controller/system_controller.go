package common_controller

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/sys"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SystemController struct {
	_conf *conf.Conf
	_co   *control_pc.ControlPCService
}

func NewSystemController(_co *control_pc.ControlPCService, _conf *conf.Conf) *SystemController {
	return &SystemController{_co: _co, _conf: _conf}
}

// @Summary Get Software Info
// @Description Get the software version information
// @Tags System
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData	  "success"
// @Failure 500 {object} schema.ResponseData	 "Server internal error"
// @Router /info [get]
func (s *SystemController) GetSoftwareInfo(c *gin.Context) {
	ver, err := version.GetVersion()
	if err != nil {
		c.Error(err)
		return

	}
	supportAlgo := make([]schema.AlgorithmInfo, 0)
	for algo, _ := range secure.AlgorithmNames {
		supportAlgo = append(supportAlgo, schema.AlgorithmInfo{AlgorithmName: secure.AlgorithmNames[algo]})
	}

	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c,
		schema.SoftwareInfo{
			Version: ver,
			Edition: version.GetEdition(),
			ServiceInfo: []schema.ServiceInfo{
				{
					ServiceName: sys.ServiceName,
				},
			},
			WorkDir:       s._conf.GetWorkdir(),
			LogLevel:      logger.GetLogLevel(),
			LogPath:       logger.GetLogPath(),
			AlgorithmInfo: supportAlgo,
			AppVersion:    "3.5.0.7",
		}))
}

// @Summary Set Power Saving Mode
// @Description Enable or disable power saving mode. This function first records the setting in the database. If the database write fails, it returns immediately. If the database write succeeds, it sets the power saving mode. A failure in setting the mode does not affect the database value.
// @Tags System
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param enable query string false "Enable power saving mode (true or false)" default(true)
// @Success 200 {object} schema.ResponseData "Power saving mode set successfully."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /power-saving [post]
func (s *SystemController) SetPowerSavingMode(c *gin.Context) {

	enable := c.DefaultQuery("enable", "true")
	if enable == "true" {
		err := s._co.SetPowerSavingMode(true)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return
	} else if enable == "false" {
		err := s._co.SetPowerSavingMode(false)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return
	}
	c.Error(exception.ErrUserParameterError)

}
