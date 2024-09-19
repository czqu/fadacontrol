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
	"fadacontrol/pkg/utils"
	"fmt"
	RMTT "github.com/czqu/rmtt-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"math"
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

func (r *RemoteService) ProtoHandler(client RMTT.Client, data, requestId *[]byte) {

	var msg = &remote_schema.RemoteMsg{}
	err := proto.Unmarshal(*data, msg)
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

	var decodeData []byte
	switch packet.EncryptionAlgorithm {
	case remote_schema.None:
		break
	case remote_schema.AESGCM128Algorithm:
		{
			salt, err := base64.StdEncoding.DecodeString(r.config.Salt)
			if err != nil {
				logger.Warn(err)
				r.PushTextRet(client, exception.ErrDecryptDataError, packet.RequestId)
				return
			}
			decodeData, err = remote_schema.DecryptData(*packet.Data, salt, r.config.Secret)

			if err != nil {
				logger.Warn(err)
				r.PushTextRet(client, exception.ErrDecryptDataError, packet.RequestId)
				return
			}
		}
	default:
		{
			r.PushTextRet(client, exception.ErrUnsupportedCryptographicAlgorithm, packet.RequestId)
		}

	}
	switch packet.DataType {
	case remote_schema.ProtoBuf:
		r.ProtoHandler(client, &decodeData, packet.RequestId)
	default:
		r.PushTextRet(client, exception.ErrParameterError, packet.RequestId)
	}

}

func (r *RemoteService) PushTextRet(client RMTT.Client, ret *exception.Exception, requestId *[]byte) {
	if len(*requestId) > 0xff {
		return
	}
	requestIdLen := uint8(len(*requestId))
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, ret.Code)
	if err != nil {
		logger.Warn(err)
		return
	}

	payload := buffer.Bytes()
	packet := &remote_schema.PayloadPacket{EncryptionAlgorithm: remote_schema.None, DataType: remote_schema.Text,
		Data: &payload, RequestIdLen: requestIdLen, RequestId: requestId}
	if client != nil {

		client.Push(packet)
	}
}

func (r *RemoteService) PushProtoRet(client RMTT.Client, encryptFlag bool, ex *exception.Exception, requestId *[]byte) {
	if len(*requestId) > 0xff {
		return
	}
	requestIdLen := uint8(len(*requestId))
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
		packet := &remote_schema.PayloadPacket{EncryptionAlgorithm: remote_schema.None, DataType: remote_schema.ProtoBuf, Data: &data}
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
	key, err := base64.StdEncoding.DecodeString(r.config.Key)
	if err != nil {
		logger.Error(err)
		return
	}
	encryptData, err := secure.EncryptAESGCM(key, data)
	if err != nil {
		logger.Error(err)
	}

	packet := &remote_schema.PayloadPacket{RequestIdLen: requestIdLen, RequestId: requestId, EncryptionAlgorithm: remote_schema.AESGCM256Algorithm, DataType: remote_schema.ProtoBuf, Data: &encryptData}

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
	url := "https://api.voidbytes.com/v1/client/client-id"

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
func (r *RemoteService) UpdateData(_c remote_schema.RemoteConfigReqDTO) error {
	c := r.config
	if err := r.db.First(&c).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return err
	}
	var servers []entity.RemoteServer

	for _, server := range _c.Server {
		if server.MsgServerUrl == "" || server.ApiServerUrl == "" {
			continue
		}
		s := entity.RemoteServer{MsgServerUrl: server.MsgServerUrl, ApiServerUrl: server.ApiServerUrl}
		servers = append(servers, s)

	}
	err := r.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&servers).Error
	if err != nil {
		logger.Errorf("failed to update database: %v", err)
	}
	if _c.Secret != "" {
		salt, err := secure.GenerateSalt(remote_schema.MaxKeyLength)
		if err != nil {
			return err
		}
		key, err := secure.GenerateArgon2IDKeyOneTime64MB4Threads(_c.Secret, salt, 5, remote_schema.MaxKeyLength)
		if err != nil {
			return err
		}
		c.Salt = base64.StdEncoding.EncodeToString(salt)
		c.Key = base64.StdEncoding.EncodeToString(key)
	}
	if _c.ClientId != "" {
		c.ClientId = _c.ClientId
	}
	c.Enable = _c.Enabled
	r.db.Save(&c)
	return nil
}
func (r *RemoteService) GetData() (*remote_schema.RemoteConfigRespDTO, error) {
	if err := r.db.First(&r.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, err
	}
	var servers []entity.RemoteServer
	err := r.db.Limit(10).Find(&servers).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	var rs []remote_schema.RemoteServer
	for _, s := range servers {
		e := remote_schema.RemoteServer{MsgServerUrl: s.MsgServerUrl, ApiServerUrl: s.ApiServerUrl}
		rs = append(rs, e)
	}
	return &remote_schema.RemoteConfigRespDTO{Enabled: r.config.Enable, ClientId: r.config.ClientId, Key: r.config.Key, Server: rs}, nil
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
func (r *RemoteService) CreatConfig() {
	err := r.db.AutoMigrate(&entity.RemoteConnectConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")

	}
	err = r.db.AutoMigrate(&entity.RemoteServer{})
	if err != nil {
		logger.Errorf("failed to migrate database")
	}
	var count int64
	r.db.Model(&entity.RemoteConnectConfig{}).Count(&count)
	if count == 0 {

		remoteConfig := entity.RemoteConnectConfig{
			Enable: false,
			Secret: utils.GenRandString(8),
		}
		r.db.Create(&remoteConfig)
	}

}
func (r *RemoteService) TestServerDelay() int64 {
	var ret int64
	ret = math.MinInt64
	opts := RMTT.NewClientOptions()
	err := r.loadConfig()
	if err != nil {
		return ret
	}
	server := entity.RemoteServer{}
	err = r.db.First(&server).Error
	if err != nil {
		return ret
	}
	if server.MsgServerUrl == "" {
		return ret
	}
	opts.AddServer(server.MsgServerUrl)
	opts.SetClientID("test-client-id")
	client := RMTT.NewClient(opts)
	client.AddPayloadHandlerLast(nil)

	now := time.Now()
	token := client.Connect()
	token.SetErrorHandler(func(err error) {
		if errors.Is(err, RMTT.RefusedNotAuthorisedErr) {
			ret = int64(time.Since(now))
		}
		logger.Warn(err)
	})
	token.Wait()

	return ret
}

func (r *RemoteService) StartService() {
	logger.Debug("starting service")
	r.CreatConfig()
	err := r.loadConfig()
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}
	if r._conf.Debug {
		r.config.Enable = true
	}
	if r.config.Enable == false {
		return
	}
	RMTT.DEBUG = logger.GetLogger()
	RMTT.ERROR = logger.GetLogger()
	RMTT.INFO = logger.GetLogger()
	RMTT.WARN = logger.GetLogger()

	opts := RMTT.NewClientOptions()
	server := entity.RemoteServer{}
	r.db.First(&server)
	if server.MsgServerUrl == "" {
		return
	}
	opts.AddServer(server.MsgServerUrl)
	if r._conf.Debug {
		opts.AddServer("tcp://127.0.0.1:3016")
	}

	opts.SetClientID(r.config.ClientId)
	logger.Debug("your client id is ", r.config.ClientId)
	client := RMTT.NewClient(opts)
	client.AddPayloadHandlerLast(r.RRFPMsgHandler)

	go func() {

		token := client.Connect()
		token.Wait()
		err = token.Error()
		if err != nil {

			logger.Warnf("connect fail%v", token.Error())
			if errors.Is(err, RMTT.RefusedNotAuthorisedErr) || errors.Is(err, RMTT.ProtocolViolationErr) {
				return
			}
			r.RestartService()
			return
		}

	}()

	r.statusLock.Lock()
	r.runStatus = true
	r.statusLock.Unlock()
	logger.Debug("done")
	<-r.done
	logger.Debug("close")
	client.Disconnect(0)

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
