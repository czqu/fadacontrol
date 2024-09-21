package common_controller

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/sys"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SysInfoController struct {
	_conf *conf.Conf
}

func NewSysInfoController(_conf *conf.Conf) *SysInfoController {
	return &SysInfoController{_conf: _conf}
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
func (s *SysInfoController) GetSoftwareInfo(c *gin.Context) {
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
		}))
}
