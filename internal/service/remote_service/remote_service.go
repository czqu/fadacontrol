package remote_service

import (
	"encoding/json"
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema/remote_schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/remote_service/rml"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/pkg/utils"
	"fmt"
	RMTT "github.com/czqu/rmtt-go"
	"gorm.io/gorm"
	"io"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type RemoteService struct {
	co         *control_pc.ControlPCService
	un         *unlock.UnLockService
	_conf      *conf.Conf
	db         *gorm.DB
	config     *entity.RemoteConnectConfig
	Client     RMTT.Client
	done       chan struct{}
	runStatus  bool
	statusLock sync.Mutex
}

func NewRemoteService(co *control_pc.ControlPCService, un *unlock.UnLockService, _conf *conf.Conf, db *gorm.DB) *RemoteService {
	return &RemoteService{co: co, un: un, _conf: _conf, db: db, done: make(chan struct{}), config: &entity.RemoteConnectConfig{}}
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
		PushRet(client, 20011)
		return
	}
	decodeData, err := remote_schema.DecryptData(packet.Data, packet.Salt, r.config.Secret)

	if err != nil || packet.DataType != remote_schema.JsonType {
		logger.Warn(err)
		PushRet(client, 20011)
		return
	}
	runtime.GC()
	var jsonData map[string]interface{}
	err = json.Unmarshal(decodeData, &jsonData)
	if err != nil {
		//todo
		logger.Error(err)
		return
	}
	typ, ok := jsonData["type"].(float64)
	if !ok {
		logger.Warn("Error: unable to read type from JSON")
	}

	var ret *exception.Exception = exception.ErrUnknownException
	switch int(typ) {
	case Unlock:
		{
			logger.Debug("recv unlock data")
			username, password, err := rml.ReadJson(string(decodeData))
			if err != nil {
				logger.Warn(err)
				PushRet(client, 20012)
				return
			}
			if username == "" || password == "" {
				PushRet(client, 10009)
				return
			}
			ret = r.un.UnlockPc(username, password)

		}
	case Standby:
		ret = r.co.Standby()
	case Shutdown:
		ret = r.co.Shutdown()
	case LockScreen:
		ret = r.co.LockWindows(true)

	default:
		logger.Warnf("not support type: %v", typ)

	}
	PushRet(client, ret.Code)

	return
}

type DoPushFun func(msg []byte)

func PushRet(client RMTT.Client, code int) {
	pushPayload := make([]byte, 1)
	pushPayload[0] = 0
	pushData := struct {
		ErrCode int         `json:"err_code"`
		Data    interface{} `json:"data"`
	}{
		ErrCode: code,
		Data:    nil,
	}
	jsonData, err := json.Marshal(pushData)
	if err != nil {
		return
	}
	pushPayload = append(pushPayload, jsonData...)
	doPush(client, pushPayload)
}
func doPush(client RMTT.Client, msg []byte) {
	if client != nil {

		client.Push(msg)
	}
}

func (r *RemoteService) IdentifyMe() {
	pushData := struct {
		ErrCode int `json:"err_code"`
		Data    struct {
			Identifier string `json:"identifier"`
		} `json:"data"`
	}{

		ErrCode: 0,
		Data: struct {
			Identifier string `json:"identifier"`
		}{
			Identifier: "unlock",
		},
	}
	_jsonData, err := json.Marshal(pushData)
	if err != nil {
		return
	}
	r.PushPacketNoSecure(_jsonData)
}
func (r *RemoteService) PushPacketNoSecure(data []byte) {
	pushPayload := make([]byte, 1)
	pushPayload[0] = 0
	pushPayload = append(pushPayload, data...)

	doPush(r.Client, pushPayload)

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
func (r *RemoteService) UpdateData(_c remote_schema.RemoteConfigDTO) error {
	if err := r.db.First(r.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return err
	}
	if r.config.ClientId != "" {
		r.config.ClientId = _c.ClientId
	}

	if r.config.Secret != "" {

		r.config.Secret = _c.Secret
	}

	r.config.Enable = _c.Enabled
	r.config.Url = _c.Url

	r.db.Save(&r.config)
	return nil
}
func (r *RemoteService) GetData() (*remote_schema.RemoteConfigDTO, error) {
	if err := r.db.First(&r.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, err
	}
	if r.config.ClientId == "" {
		clientId, err := r.GetClientId()
		if err == nil {
			r.config.ClientId = clientId
		}
	}

	if r.config.Secret == "" {
		secret := utils.GenRandString(8)

		r.config.Secret = secret

	}
	r.db.Save(&r.config)
	return &remote_schema.RemoteConfigDTO{Enabled: r.config.Enable, ClientId: r.config.ClientId, Secret: r.config.Secret, Url: r.config.Url}, nil
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
		return
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
	opts.SetServer(r.config.Url)
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
	if r.config.Enable == false {
		return
	}
	RMTT.DEBUG = logger.GetLogger()
	RMTT.ERROR = logger.GetLogger()
	RMTT.INFO = logger.GetLogger()
	RMTT.WARN = logger.GetLogger()

	opts := RMTT.NewClientOptions()
	opts.SetServer(r.config.Url)

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

		r.IdentifyMe()
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
