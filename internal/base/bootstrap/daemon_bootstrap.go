package bootstrap

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"net"
	"strconv"
	"time"
)

type DaemonBootstrap struct {
	ChanGroup *conf.ChanGroup
	_conf     *conf.Conf
	done      chan interface{}
}

func NewDaemonBootstrap(g *conf.ChanGroup, _conf *conf.Conf) *DaemonBootstrap {
	return &DaemonBootstrap{ChanGroup: g, _conf: _conf, done: make(chan interface{})}
}

func (d *DaemonBootstrap) Start() {
	port := 2095
	host := "127.0.0.1"
	addr := host + ":" + strconv.Itoa(port)
	go d.connectToServer(addr)
}
func (d *DaemonBootstrap) Stop() {
	d.done <- struct{}{}
}

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 16 * time.Second
)

func (d *DaemonBootstrap) connectToServer(addr string) {
	backoff := initialBackoff

NewConn:
	for {
		conn, err := net.Dial("tcp", addr)
		select {
		case <-d.done:
			return
		case <-d.ChanGroup.InternalCommandSend: //clean
			logger.Debug("receive command    ")
			break
		default:
			break
		}
		if err != nil || conn == nil {
			logger.Debugf("Error connecting to server: %v\n", err)

			// Wait for the backoff time and try again
			time.Sleep(backoff)

			// Increase the backoff time until maxBackoff is reached
			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}
		tcpConn := conn.(*net.TCPConn)
		err = tcpConn.SetKeepAlive(true)
		if err != nil {
			logger.Warn("Error setting keep-alive:", err)
		}
		cmd := schema.InternalCommand{CommandType: schema.Hello, Data: nil}
		msg, err := json.Marshal(cmd)
		if err != nil {
			logger.Error(err)
			return
		}
		packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
		data, err := packet.Pack()
		if err != nil {
			logger.Error(err)
			return
		}
		conn.Write(data)

		// Reset the backoff time after a successful connection
		backoff = initialBackoff
		keepAlive := 30 * time.Second
		if d._conf.Debug {
			keepAlive = 15 * time.Second
		}
		logger.Debug("connected")
		for {
			select {
			case <-d.done:
				return
			case msg := <-d.ChanGroup.InternalCommandSend:
				packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
				logger.Debug("receive command    ")
				data, err := packet.Pack()
				if err != nil {
					logger.Error(err)
					break
				}
				_, err = conn.Write(data)
				if err != nil {
					logger.Debug(err)
					goto NewConn
				}
			case <-time.After(keepAlive):
				cmd := schema.InternalCommand{CommandType: schema.KeepLive, Data: nil}
				logger.Debug("send keep live to server")
				msg, err := json.Marshal(cmd)
				packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
				data, err := packet.Pack()
				if err != nil {
					logger.Error(err)
					break
				}
				_, err = conn.Write(data)
				if err != nil {
					logger.Debug(err)
					goto NewConn
				}
			}
		}

	}
}
