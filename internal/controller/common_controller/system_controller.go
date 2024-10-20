package common_controller

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/update_service"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/syncer"
	"fadacontrol/pkg/sys"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SystemController struct {
	_conf *conf.Conf
	_co   *control_pc.ControlPCService
	_up   *update_service.UpdateService
}

func NewSystemController(_co *control_pc.ControlPCService, _conf *conf.Conf, _up *update_service.UpdateService) *SystemController {
	return &SystemController{_co: _co, _conf: _conf, _up: _up}
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
	ver := version.GetVersion()

	supportAlgo := make([]schema.AlgorithmInfo, 0)
	for algo, _ := range secure.AlgorithmNames {
		supportAlgo = append(supportAlgo, schema.AlgorithmInfo{AlgorithmName: secure.AlgorithmNames[algo]})
	}

	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c,
		schema.SoftwareInfo{
			Version:    ver,
			BuildInfo:  version.GetBuildInfo(),
			Edition:    string(version.GetEdition()),
			AppVersion: version.GetVersionName(),
			ServiceInfo: []schema.ServiceInfo{
				{
					ServiceName: sys.ServiceName,
				},
			},
			WorkDir:       s._conf.GetWorkdir(),
			LogLevel:      logger.GetLogLevel(),
			LogPath:       logger.GetLogPath(),
			AlgorithmInfo: supportAlgo,
			AuthorEmail:   version.AuthorEmail,
		}))
}

// @Summary Set Power Saving Mode
// @Description Enable or disable power saving mode. This function first records the setting in the database. If the database write fails, it returns immediately. If the database write succeeds, it sets the power saving mode. A failure in setting the mode does not affect the database value.
// @Tags System
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param mode query string false "Enable power saving mode (enable or disable or auto)" default(auto)
// @Success 200 {object} schema.ResponseData "Power saving mode set successfully."
// @Failure 400 {object} schema.ResponseData "Invalid request parameters."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /power-saving [post]
func (s *SystemController) SetPowerSavingMode(c *gin.Context) {

	enable := c.DefaultQuery("mode", "auto")
	if enable == "enable" || enable == "auto" {
		err := s._co.SetPowerSavingMode(true)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccess(c))
		return
	} else if enable == "disable" {
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

// @Summary Get System Logs
// @Description Stream the system logs in real-time. This endpoint opens a connection to the log buffer and sends log entries as they are generated. The connection remains open until explicitly closed or an error occurs. If the buffer is not available, it returns an error response.
// @Tags System
// @Security ApiKeyAuth
// @Produce text/event-stream
// @Param module path string true "Specify the module to retrieve logs from (must be 'service')"
// @Success 200 {string} string "Stream of system logs."
// @Failure 400 {object} schema.ResponseData "Invalid module specified."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /logs/{module} [get]
func (s *SystemController) GetLog(c *gin.Context) {
	module := c.Param("module")
	if module != "service" {
		c.Error(exception.ErrUserParameterError)
		return
	}
	log := logger.GetLogger()
	if log != nil {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		ctx := c.Request.Context()

		w := syncer.AddResponseSyncer(c.Writer)
		id := log.AddReader(w)

		<-ctx.Done()
		log.RemoveWriter(id)

	}
	c.JSON(http.StatusOK, controller.GetGinError(c, exception.ErrUnknownException))
}

// @Summary Check for System Updates
// @Description Check if there are any updates available for the system. This endpoint returns the latest available update details, including the version, update URL, and release notes.
// @Tags System
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} schema.ResponseData "Successfully retrieved update information."
// @Failure 500 {object} schema.ResponseData "Internal Server Error"
// @Router /info/check_update [get]
func (s *SystemController) CheckUpdate(c *gin.Context) {
	ret, err := s._up.CheckUpdate()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccessWithData(c, ret))
}
