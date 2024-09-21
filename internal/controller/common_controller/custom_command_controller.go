package common_controller

import (
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema/custom_command_schema"
	"fadacontrol/internal/service/custom_command_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type chanGroup struct {
	stdout *custom_command_schema.CustomWriter
	stderr *custom_command_schema.CustomWriter
}
type CustomCommandController struct {
	_conf    *conf.Conf
	service  *custom_command_service.CustomCommandService
	cmdCache map[string]*chanGroup
}

func NewCustomCommandController(_conf *conf.Conf, service *custom_command_service.CustomCommandService) *CustomCommandController {
	return &CustomCommandController{_conf: _conf, service: service, cmdCache: make(map[string]*chanGroup)}
}

func (d *CustomCommandController) Execute(c *gin.Context) {
	var dto custom_command_schema.CustomCommandReq
	err := c.ShouldBind(&dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	cmds, err := d.service.ReadConfig(dto.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	cmd, ok := cmds[dto.Name]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("command not found")})
	}
	_uuid, err := uuid.NewRandom()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	commandID := _uuid.String()

	resp := custom_command_schema.CustomCommandResp{Id: commandID}
	stdout := custom_command_schema.NewCustomWriter()
	stderr := custom_command_schema.NewCustomWriter()
	g := &chanGroup{stdout: stdout, stderr: stderr}
	d.cmdCache[commandID] = g
	err = d.service.ExecuteCommand(cmd, stdout, stderr)

	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	c.JSON(http.StatusOK, resp)

}
func (d *CustomCommandController) ExecResult(c *gin.Context) {
	commandID := c.Param("id")
	g, ok := d.cmdCache[commandID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}
	stdoutChan := g.stdout.Ch
	stderrChan := g.stderr.Ch
	defer delete(d.cmdCache, commandID)
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	ctx := c.Request.Context()

	for {
		select {
		case msg := <-stdoutChan:
			c.Writer.Write(msg)
			c.Writer.Flush()
		case msg := <-stderrChan:
			c.Writer.Write(msg)
			c.Writer.Flush()
		case <-ctx.Done():
			logger.Debug("Client closed connection")
			//delete(d.cmdCache, commandID)
			return
		}
	}

}
