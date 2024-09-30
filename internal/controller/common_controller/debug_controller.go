package common_controller

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fadacontrol/pkg/goroutine"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var (
	activeConnections = 0
	maxConnections    = 2
	mu                sync.Mutex
)

// @Summary	Ping
// @Produce	json
// @Success	200		"success"
// @Router		/ping [get]
func Ping(c *gin.Context) {
	query := c.Query("type")
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
	c.JSON(http.StatusOK, controller.GetGinSuccess(c))
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
