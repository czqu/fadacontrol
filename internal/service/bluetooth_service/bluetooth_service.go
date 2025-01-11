package bluetooth_service

import (
	"context"
	"encoding/binary"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema/bluetooth_schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/sys/bluetooth"
	"fadacontrol/pkg/utils"
	"fmt"
	"io"
	"time"

	"net"
	"sync"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type bluetoothData struct {
	Size uint16
	Data []byte
}
type BluetoothService struct {
	co                 *control_pc.ControlPCService
	u                  *unlock.UnLockService
	ctx                context.Context
	db                 *gorm.DB
	listener           net.Listener
	bluetoothCtx       context.Context
	bluetoothCtxCancel context.CancelFunc
	bluetoothCtxLock   sync.Mutex
}

func NewBluetoothService(ctx context.Context, db *gorm.DB, co *control_pc.ControlPCService, u *unlock.UnLockService) *BluetoothService {
	return &BluetoothService{co: co, u: u, ctx: ctx, db: db}
}

func (r *BluetoothService) StartService() error {
	r.bluetoothCtx, r.bluetoothCtxCancel = context.WithCancel(r.ctx)

	config := &bluetooth.Config{ServiceInstanceName: "RemoteFingerprint Service ", Comment: "RemoteFingerprint Service "}

	serviceClassId := bluetooth.GUID{
		Data1: 0x4E5877C0,
		Data2: 0x8297,
		Data3: 0x4AAE,
		Data4: [8]byte{0xB7, 0xBD, 0x73, 0xA8, 0xCB, 0xC1, 0xED, 0xAF},
	}
	var err error
	r.listener, err = bluetooth.Listen(serviceClassId, config)
	if err != nil {
		logger.Errorf("Failed to listen: %v", err)
		return err
	}
	defer r.listener.Close()
	for {
		logger.Infof("Waiting for a ble connection...")
		conn, err := r.listener.Accept()
		if err != nil {
			logger.Errorf("Failed to accept connection: %v", err)
			return err
		}
		select {
		case <-r.bluetoothCtx.Done():
			return nil
		default:
		}

		goroutine.RecoverGO(
			func() {
				r.handleConnection(conn)
			})
	}
}

// sendPacket 是一个辅助函数，用于将 Protobuf 消息封包（长度头+数据）并发送
func (r *BluetoothService) sendPacket(conn net.Conn, packet *bluetooth_schema.BluetoothPacket) error {
	// 1. 序列化 Protobuf 消息
	respData, err := proto.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal response packet: %w", err)
	}

	// 2. 获取数据长度，并以大端序写入2字节的长度头
	dataSize := uint16(len(respData))
	if err := binary.Write(conn, binary.BigEndian, dataSize); err != nil {
		return fmt.Errorf("failed to write response size: %w", err)
	}

	// 3. 写入实际的数据
	_, err = conn.Write(respData)
	if err != nil {
		return fmt.Errorf("failed to write response data: %w", err)
	}

	return nil
}

// handleConnection 处理单个蓝牙客户端连接
func (r *BluetoothService) handleConnection(conn net.Conn) {
	defer conn.Close()
	logger.Debugf("Client %s connected", conn.RemoteAddr().String())

	// 循环读取客户端发送的数据包
	for {
		// 1. 读取2字节的数据包长度
		var dataSize uint16
		err := binary.Read(conn, binary.BigEndian, &dataSize)
		if err != nil {
			if err == io.EOF {
				logger.Errorf("Client %s actively disconnects", conn.RemoteAddr().String())
			} else {
				logger.Errorf("Read data length failed: %v", err)
			}
			return // 结束该客户端的处理
		}

		// 2. 根据长度读取完整的数据包内容
		data := make([]byte, dataSize)
		_, err = io.ReadFull(conn, data)
		if err != nil {
			logger.Errorf("Failed to read the data body: %v", err)
			return
		}

		// 3. 将字节数据反序列化为 Protobuf 对象
		var reqPacket bluetooth_schema.BluetoothPacket
		if err := proto.Unmarshal(data, &reqPacket); err != nil {
			logger.Errorf("Protobuf DESERIALIZATION FAILED: %v", err)
			continue // 继续等待下一个包
		}

		// 4. 处理请求并准备响应
		// 从请求包中提取请求体
		reqPayload, ok := reqPacket.Payload.(*bluetooth_schema.BluetoothPacket_Request)
		if !ok {
			logger.Errorf("Received an invalid package type, not a Request type")
			continue
		}
		request := reqPayload.Request

		var code int32
		var msg string

		// 根据请求类型分发到不同的业务逻辑函数
		switch request.GetType() {
		case bluetooth_schema.MsgType_Unlock:
			// 从 oneof body 中获取具体的 UnlockMsg
			unlockMsg := request.GetUnlockMsg()
			if unlockMsg != nil {
				// 调用内部的解锁函数
				code, msg = r.handleUnlock(unlockMsg.GetUsername(), unlockMsg.GetPassword())
			} else {
				code = int32(exception.ErrUserParameterError.Code) // Bad Request
				msg = "unlockMsg is missing from the request body"
			}

		case bluetooth_schema.MsgType_Shutdown:
			shutdownMsg := request.GetShutdownMsg()
			if shutdownMsg != nil {
				// 调用内部的关机函数
				code, msg = r.handleShutdown(shutdownMsg.GetType())
			} else {
				code = int32(exception.ErrUserParameterError.Code)
				msg = "shutdownMsg is missing from the request body"
			}

		case bluetooth_schema.MsgType_Standby:
			code, msg = r.handleStandby()
		case bluetooth_schema.MsgType_LockScreen:
			code, msg = r.lockScreen()
		default:
			code = int32(exception.ErrUserParameterError.Code)
			msg = fmt.Sprintf("unsupported request type: %v", request.GetType())
		}

		// 5. 构建响应包
		respPacket := &bluetooth_schema.BluetoothPacket{
			// 核心：将请求的 sequence_id 原样返回
			SequenceId: reqPacket.GetSequenceId(),
			Payload: &bluetooth_schema.BluetoothPacket_Response{
				Response: &bluetooth_schema.Response{
					Code: code,
					Msg:  msg,
				},
			},
		}

		// 6. 发送响应包
		if err := r.sendPacket(conn, respPacket); err != nil {
			logger.Errorf("Failed to send response: %v", err)
			return
		}
		logger.Errorf("Requests Responded (ID: %d)", reqPacket.GetSequenceId())
	}
}

func (r *BluetoothService) StopService() error {
	if r.bluetoothCtx != nil {
		r.bluetoothCtxLock.Lock()
		r.bluetoothCtxCancel()
		r.bluetoothCtxLock.Unlock()
	}
	return r.listener.Close()

}

func (r *BluetoothService) GetBluetoothConfig() (*bluetooth_schema.BluetoothSchema, error) {
	var config entity.BluetoothConfig
	err := r.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &bluetooth_schema.BluetoothSchema{Enabled: config.Enabled}, nil
}

func (r *BluetoothService) PatchBluetoothConfig(m *map[string]interface{}) error {

	var config entity.BluetoothConfig
	err := r.db.First(&config).Error
	if err != nil {
		return err
	}
	if err := r.db.Model(&config).Updates(*m).Error; err != nil {
		return err
	}
	return nil

}

func (r *BluetoothService) RestartBoothService() error {
	r.StopService()
	return r.StartService()
}

func (r *BluetoothService) handleUnlock(username string, password string) (int32, string) {
	e := r.u.UnlockPc(username, password)
	if e != nil {
		return int32(e.Code), e.Msg
	}
	return exception.UnknownError, "unknown error"
}

func (r *BluetoothService) lockScreen() (int32, string) {
	_conf := utils.GetValueFromContext(r.ctx, constants.ConfKey, conf.NewDefaultConf())
	var ret error
	if _conf.StartMode == conf.ServiceMode {
		ret = r.co.LockWindows(true)
	} else {
		ret = r.co.LockWindows(false)
	}
	if ret == nil {
		return exception.UnknownError, "unknown error"
	}

	return int32(exception.ErrUserParameterError.Code), ret.Error()

}

var delaySec = 5

func (r *BluetoothService) handleShutdown(getType bluetooth_schema.ShutdownType) (int32, string) {

	goroutine.RecoverGO(func() {
		time.Sleep(time.Duration(delaySec) * time.Second)
		e := r.co.Shutdown(protoTypeToShutdownType(getType))
		if exception.ErrSuccess.NotEqual(e) {
			logger.Errorf("Shutdown failed: %v", e)
		}
	})

	return int32(exception.ErrSuccess.Code), "success"
}
func (r *BluetoothService) handleStandby() (int32, string) {
	goroutine.RecoverGO(func() {
		time.Sleep(time.Duration(delaySec) * time.Second)
		e := r.co.Standby()
		if exception.ErrSuccess.NotEqual(e) {
			logger.Errorf("Standby failed: %v", e)
		}
	})
	return int32(exception.ErrSuccess.Code), "success"
}
func protoTypeToShutdownType(getType bluetooth_schema.ShutdownType) sys.ShutdownType {
	switch getType {
	case bluetooth_schema.ShutdownType_E_FORCE_REBOOT:
		return sys.S_E_FORCE_REBOOT
	case bluetooth_schema.ShutdownType_E_LOGOFF:
		return sys.S_E_LOGOFF
	case bluetooth_schema.ShutdownType_E_FORCE_SHUTDOWN:
		return sys.S_E_FORCE_SHUTDOWN
	case bluetooth_schema.ShutdownType_EWX_FORCE_POWEROFF:
		return sys.S_EWX_FORCE_POWEROFF
	case bluetooth_schema.ShutdownType_EWX_REBOOT:
		return sys.S_EWX_REBOOT
	case bluetooth_schema.ShutdownType_EWX_REBOOT_RESTARTAPPS:
		return sys.S_EWX_REBOOT_RESTARTAPPS
	case bluetooth_schema.ShutdownType_EWX_POWEROFF:
		return sys.S_EWX_POWEROFF

	case bluetooth_schema.ShutdownType_EWX_HYBRID_SHUTDOWN:
		return sys.S_EWX_HYBRID_SHUTDOWN
	case bluetooth_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_FORCE:
		return sys.S_EWX_HYBRID_SHUTDOWN_FORCE
	case bluetooth_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_RESTARTAPPS:
		return sys.S_EWX_HYBRID_SHUTDOWN_RESTARTAPPS
	case bluetooth_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS:
		return sys.S_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS
	case bluetooth_schema.ShutdownType_EWX_SHUTDOWN_RESTARTAPPS:
		return sys.S_EWX_SHUTDOWN_RESTARTAPPS
	case bluetooth_schema.ShutdownType_EWX_SHUTDOWN:
		return sys.S_EWX_SHUTDOWN
	case bluetooth_schema.ShutdownType_EWX_FORCE_REBOOT_RESTARTAPPS:
		return sys.S_EWX_FORCE_REBOOT_RESTARTAPPS

	default:
		return sys.Unknown
	}

}
