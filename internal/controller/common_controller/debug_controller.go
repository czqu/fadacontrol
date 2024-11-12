package common_controller

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
	"time"
)

type DebugController struct {
	_im   *internal_service.InternalMasterService
	_conf *conf.Conf
}

func NewDebugController(im *internal_service.InternalMasterService, _conf *conf.Conf) *DebugController {
	return &DebugController{_im: im, _conf: _conf}
}

var (
	activeConnections = 0
	maxConnections    = 2
	mu                sync.Mutex
)

// @Summary	Ping
// @Produce	json
// @Param	type	query	string	false	"type" Enums(ws,pairing)
// @Success	200		"success"
// @Failure	500		"error"
// @Router		/ping [get]
func (d *DebugController) Ping(c *gin.Context) {
	query := c.DefaultQuery("type", "")
	if query == "ws" {
		mu.Lock()
		if activeConnections >= maxConnections {
			mu.Unlock()
			c.Error(exception.ErrUserTooManyRequests.SetMsg("too many connections"))
			return
		}
		activeConnections++
		mu.Unlock()
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("upgrade error: %v", err)
			return
		}
		defer func() {
			conn.Close()
			mu.Lock()
			activeConnections--
			mu.Unlock()
		}()
		goroutine.RecoverGO(func() {
			time.Sleep(5 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("write error: %v", err)

			}
		})
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				logger.Error("Read error:", err)
				break
			}

			if messageType == websocket.PingMessage {
				logger.Debug("Received Ping")
				if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
					logger.Error("Write Pong error:", err)
					break
				}
			} else {
				logger.Debug("Received message: ", message)
			}

		}

	}
	if query == "full_status" {
		if d._conf.StartMode == conf.ServiceMode && !d._im.HasClient() {
			c.Error(exception.ErrSystemServiceNotFullyStarted)
			return
		}

	}
	if query == "pairing" {

		hostname, err := os.Hostname()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, controller.GetGinSuccessWithData(
			c,
			map[string]string{
				"hostname": hostname,
			}))
		return
	}
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
	return
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
