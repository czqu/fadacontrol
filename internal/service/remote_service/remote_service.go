package remote_service

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema/remote_schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/sys"
	"fmt"
	RMTT "github.com/czqu/rmtt-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"io"
	"net/http"
	"sync"
	"time"
)

type RemoteService struct {
	co         *control_pc.ControlPCService
	un         *unlock.UnLockService
	_conf      *conf.Conf
	db         *gorm.DB
	config     entity.RemoteConnectConfig
	Client     RMTT.Client
	done       chan struct{}
	runStatus  bool
	statusLock sync.Mutex
}

const DefaultGenKeyLength = 48

func NewRemoteService(co *control_pc.ControlPCService, un *unlock.UnLockService, _conf *conf.Conf, db *gorm.DB) *RemoteService {
	return &RemoteService{co: co, un: un, _conf: _conf, db: db, done: make(chan struct{}), config: entity.RemoteConnectConfig{}}
}

const (
	Unknown = iota
	Unlock
	LockScreen
	Shutdown
	Standby
	Reboot
	PowerOn

	END
)

func (r *RemoteService) ProtoHandler(client RMTT.Client, data, requestId []byte) {

	var msg = &remote_schema.RemoteMsg{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		logger.Warn(err)
		r.PushProtoRet(client, true, exception.ErrDeserializationError, requestId)
	}
	switch msg.Type {
	case remote_schema.MsgType_Unknown:
		r.PushProtoRet(client, true, exception.ErrParameterError, requestId)
	case remote_schema.MsgType_Unlock:
		{
			unlockMsg := msg.GetUnlockMsg()
			if unlockMsg == nil {
				r.PushProtoRet(client, true, exception.ErrParameterError, requestId)
				return
			}
			ret := r.un.UnlockPc(unlockMsg.Username, unlockMsg.Password)
			r.PushProtoRet(client, true, ret, requestId)
		}
	case remote_schema.MsgType_LockScreen:
		{
			ret := r.co.LockWindows(true)
			r.PushProtoRet(client, true, ret, requestId)
		}
	case remote_schema.MsgType_Shutdown:
		{
			shutdownMsg := msg.GetShutdownMsg()
			if shutdownMsg == nil {
				r.PushProtoRet(client, true, exception.ErrParameterError, requestId)
				return
			}
			shutdownTpe := sys.ProtoTypeToShutdownType(shutdownMsg.Type)
			ret := r.co.Shutdown(shutdownTpe)
			r.PushProtoRet(client, true, ret, requestId)
		}
	case remote_schema.MsgType_Standby:
		{
			ret := r.co.Standby()
			r.PushProtoRet(client, true, ret, requestId)
		}

	}

	return
}

func (r *RemoteService) RRFPMsgHandler(client RMTT.Client, msg RMTT.Message) {
	r.Client = client
	dataSlice := msg.Payload()
	if len(dataSlice) == 0 {
		return
	}
	packet := &remote_schema.PayloadPacket{}
	err := packet.Unpack(dataSlice) //DecodeAesPack(r.config.Secret, dataSlice)
	if err != nil {
		logger.Warn(err)
		r.PushTextRet(client, exception.ErrControlPacketParseError, packet.RequestId)
		return
	}

	key, err := secure.DecodeBase58Key(r.config.SecurityKey)
	if err != nil {
		logger.Warn(err)
		r.PushTextRet(client, exception.ErrDecryptDataError, packet.RequestId)
		return

	}
	decrpyData, err := secure.DecryptData(packet.EncryptionAlgorithm, packet.Data, key)
	if err != nil {
		logger.Warn(err)
		r.PushTextRet(client, exception.ErrDecryptDataError, packet.RequestId)
		return
	}
	switch packet.DataType {
	case remote_schema.ProtoBuf:
		r.ProtoHandler(client, decrpyData, packet.RequestId)
	default:
		r.PushTextRet(client, exception.ErrParameterError, packet.RequestId)
	}

}

func (r *RemoteService) PushTextRet(client RMTT.Client, ret *exception.Exception, requestId []byte) {
	if len(requestId) > 0xff {
		return
	}
	requestIdLen := uint8(len(requestId))
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, ret.Code)
	if err != nil {
		logger.Warn(err)
		return
	}

	payload := buffer.Bytes()
	packet := &remote_schema.PayloadPacket{EncryptionAlgorithm: secure.None, DataType: remote_schema.Text,
		Data: payload, RequestIdLen: requestIdLen, RequestId: requestId}
	if client != nil {

		client.Push(packet)
	}
}

func (r *RemoteService) PushProtoRet(client RMTT.Client, encryptFlag bool, ex *exception.Exception, requestId []byte) {
	if len(requestId) > 0xff {
		return
	}
	requestIdLen := uint8(len(requestId))
	msg := &remote_schema.RemoteMsg{
		Type:      remote_schema.MsgType_CommonResponse,
		Timestamp: timestamppb.New(time.Now()),
		MsgBody: &remote_schema.RemoteMsg_ResponseMsg{
			ResponseMsg: &remote_schema.CommonResponseMsg{
				Code: int32(ex.Code),
				Msg:  ex.Msg,
			},
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		logger.Error(err)
		return
	}
	if !encryptFlag {
		packet := &remote_schema.PayloadPacket{EncryptionAlgorithm: secure.None, DataType: remote_schema.ProtoBuf, Data: data}
		ret, err := packet.Pack()
		if err != nil {
			logger.Error(err)
			return
		}
		if client != nil {
			client.Push(ret)
		}
		return
	}
	key, err := base64.StdEncoding.DecodeString(r.config.SecurityKey)
	if err != nil {
		logger.Error(err)
		return
	}
	encryptData, err := secure.EncryptData(secure.AESGCM192Algorithm, data, key)
	if err != nil {
		logger.Error(err)
	}

	packet := &remote_schema.PayloadPacket{RequestIdLen: requestIdLen, RequestId: requestId, EncryptionAlgorithm: secure.AESGCM192Algorithm, DataType: remote_schema.ProtoBuf, Data: encryptData}

	ret, err := packet.Pack()
	if err != nil {
		logger.Error(err)
		return
	}
	if client != nil {

		client.Push(ret)
	}
}

type Response struct {
	Code string `json:"code"`
	Data string `json:"data"`
}

func (r *RemoteService) GetClientId() (string, error) {
	url := r.config.ApiServerUrl + "/v1/client/client-id"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	if result.Code != "0" {
		return "", fmt.Errorf("unexpected response code: %s", result.Code)
	}

	return result.Data, nil
}

//	func (r *RemoteService) UpdateData(_c remote_schema.RemoteConfigReqDTO) error {
//		return nil
//
// c := r.config
//
//	if err := r.db.First(&c).Error; err != nil {
//		logger.Errorf("failed to find database: %v", err)
//		return err
//	}
//
// var servers []entity.RemoteServer
//
//	for _, server := range _c.Server {
//		if server.MsgServerUrl == "" || server.ApiServerUrl == "" {
//			continue
//		}
//		s := entity.RemoteServer{MsgServerUrl: server.MsgServerUrl, ApiServerUrl: server.ApiServerUrl}
//		servers = append(servers, s)
//
// }
//
//	err := r.db.Clauses(clause.OnConflict{
//		DoNothing: true,
//	}).Create(&servers).Error
//
//	if err != nil {
//		logger.Errorf("failed to update database: %v", err)
//	}
//
//	if _c.Secret != "" {
//		salt, err := secure.GenerateSalt(remote_schema.MaxKeyLength)
//		if err != nil {
//			return err
//		}
//		key, err := secure.GenerateArgon2IDKeyOneTime64MB4Threads(_c.Secret, salt, 5, remote_schema.MaxKeyLength)
//		if err != nil {
//			return err
//		}
//		c.Salt = base64.StdEncoding.EncodeToString(salt)
//		c.Key = base64.StdEncoding.EncodeToString(key)
//	}
//
//	if _c.ClientId != "" {
//		c.ClientId = _c.ClientId
//	}
//
// c.Enable = _c.Enabled
// r.db.Save(&c)
// return nil
// }
func (r *RemoteService) GetConfig() (*remote_schema.RemoteConnectConfigResponse, error) {

	var config entity.RemoteConnectConfig
	if err := r.db.First(&config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	var remoteMsgServer []entity.RemoteMsgServer
	err := r.db.Where(&entity.RemoteMsgServer{RemoteConnectConfigId: config.ID}).Limit(10).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	servers := make([]string, 0)
	for _, server := range remoteMsgServer {
		servers = append(servers, server.MsgServerUrl)
	}

	return &remote_schema.RemoteConnectConfigResponse{
		ClientId:       config.ClientId,
		Enable:         config.Enable,
		SecurityKey:    config.SecurityKey,
		TimeStampCheck: config.TimeStampCheck,
		ApiServerUrl:   config.ApiServerUrl,
		MsgServerUrls:  servers,
	}, nil
}

func (r *RemoteService) UpdateRemoteConnectConfig(data *remote_schema.RemoteConnectConfigRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var config entity.RemoteConnectConfig

		if err := tx.First(&config).Error; err != nil {
			//todo
			return exception.ErrResourceNotFound
		}

		config.Enable = data.Enable
		config.ClientId = data.ClientId
		config.TimeStampCheck = data.TimeStampCheck
		config.ApiServerUrl = data.ApiServerUrl

		if err := tx.Save(&config).Error; err != nil {
			return err
		}

		for _, server := range data.MsgServerUrls {
			var msgServer entity.RemoteMsgServer
			msgServer.MsgServerUrl = server
			msgServer.RemoteConnectConfigId = config.ID

			if err := tx.Save(&msgServer).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
func (r *RemoteService) PatchRemoteConnectConfig(content map[string]interface{}) error {

	var config entity.RemoteConnectConfig
	if err := r.db.First(&config).Error; err != nil {
		return exception.ErrResourceNotFound
	}

	if err := r.db.Model(&config).Updates(content).Error; err != nil {
		return err
	}

	return nil

}
func (r *RemoteService) RestartService() error {
	r.StopService()
	r.StartService()
	logger.Debug("restarting service")
	return nil
}

func (r *RemoteService) loadConfig() error {
	if err := r.db.First(&r.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
	}
	if r.config.ClientId == "" {
		clientId, err := r.GetClientId()
		if err != nil {
			return err
		}
		r.config.ClientId = clientId

	}
	r.db.Save(&r.config)
	return nil
}

func (r *RemoteService) TestServerDelay() int64 {
	return 0
	//var ret int64
	//ret = math.MinInt64
	//opts := RMTT.NewClientOptions()
	//err := r.loadConfig()
	//if err != nil {
	//	return ret
	//}
	//server := entity.RemoteServer{}
	//err = r.db.First(&server).Error
	//if err != nil {
	//	return ret
	//}
	//if server.MsgServerUrl == "" {
	//	return ret
	//}
	//opts.AddServer(server.MsgServerUrl)
	//opts.SetClientID("test-client-id")
	//client := RMTT.NewClient(opts)
	//client.AddPayloadHandlerLast(nil)
	//
	//now := time.Now()
	//token := client.Connect()
	//token.SetErrorHandler(func(err error) {
	//	if errors.Is(err, RMTT.RefusedNotAuthorisedErr) {
	//		ret = int64(time.Since(now))
	//	}
	//	logger.Warn(err)
	//})
	//token.Wait()
	//
	//return ret
}

func (r *RemoteService) StartService() {
	return
	//logger.Debug("starting service")
	//
	//err := r.loadConfig()
	//if err != nil {
	//	logger.Errorf("failed to load config: %v", err)
	//	return
	//}
	//if r._conf.Debug {
	//	r.config.Enable = true
	//}
	//if r.config.Enable == false {
	//	return
	//}
	//RMTT.DEBUG = logger.GetLogger()
	//RMTT.ERROR = logger.GetLogger()
	//RMTT.INFO = logger.GetLogger()
	//RMTT.WARN = logger.GetLogger()
	//
	//opts := RMTT.NewClientOptions()
	//server := entity.RemoteServer{}
	//r.db.First(&server)
	//if server.MsgServerUrl == "" {
	//	return
	//}
	//opts.AddServer(server.MsgServerUrl)
	//if r._conf.Debug {
	//	opts.AddServer("tcp://127.0.0.1:3016")
	//}
	//
	//opts.SetClientID(r.config.ClientId)
	//logger.Debug("your client id is ", r.config.ClientId)
	//client := RMTT.NewClient(opts)
	//client.AddPayloadHandlerLast(r.RRFPMsgHandler)
	//
	//go func() {
	//
	//	token := client.Connect()
	//	token.Wait()
	//	err = token.Error()
	//	if err != nil {
	//
	//		logger.Warnf("connect fail%v", token.Error())
	//		if errors.Is(err, RMTT.RefusedNotAuthorisedErr) || errors.Is(err, RMTT.ProtocolViolationErr) {
	//			return
	//		}
	//		r.RestartService()
	//		return
	//	}
	//
	//}()
	//
	//r.statusLock.Lock()
	//r.runStatus = true
	//r.statusLock.Unlock()
	//logger.Debug("done")
	//<-r.done
	//logger.Debug("close")
	//client.Disconnect(0)

}
func (r *RemoteService) StopService() error {
	if r.runStatus == false {
		return nil
	}
	r.statusLock.Lock()
	defer r.statusLock.Unlock()

	logger.Debug("stopping service")
	r.done <- struct{}{}
	r.runStatus = false
	return nil
}
