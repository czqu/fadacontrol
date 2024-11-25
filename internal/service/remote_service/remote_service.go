package remote_service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	_ "fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/user_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"fadacontrol/internal/schema/remote_schema"
	"fadacontrol/internal/schema/remote_schema/rmtt_msg"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/unlock"

	"fmt"
	RMTT "github.com/czqu/rmtt-go"
	"google.golang.org/protobuf/proto"

	"gorm.io/gorm"
	"io"
)

type RemoteService struct {
	co           *control_pc.ControlPCService
	un           *unlock.UnLockService
	ctx          context.Context
	remoteCtx    context.Context
	remoteCancel context.CancelFunc
	db           *gorm.DB

	enable             bool
	accessToken        string
	accessKey          string
	accessSecret       string
	apiServerUrl       string
	defaultApiServerId uint
	clientId           string
	msgServers         []entity.RemoteRmttServers
	credential         entity.Credential
	lastError          error
	client             RMTT.Client
	waitMap            map[string]chan *remote_schema.PayloadPacket
	waitMapLock        sync.RWMutex
}

const DefaultGenKeyLength = 48

func NewRemoteService(userService *user_service.UserService, co *control_pc.ControlPCService, un *unlock.UnLockService, ctx context.Context, db *gorm.DB) *RemoteService {
	return &RemoteService{co: co, un: un, ctx: ctx, db: db}
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

func (r *RemoteService) generateMetaSignatureString(data *rmtt_msg.MetaData) (string, error) {
	var ret = ""
	ret = ret + utils.ProtoNumberToString(data.GetMsgType().Number()) + "\n"
	ret = ret + data.Nonce + "\n"
	ret = ret + utils.Int64ToString(data.GetTimestamp()) + "\n"
	if data.GetKey() != "" {
		ret = ret + data.GetKey() + "\n"
	}
	ret = ret + utils.ProtoNumberToString(data.GetSignatureMethod().Number()) + "\n"
	if data.GetBigDataMd5() != "" {
		ret = ret + data.BigDataMd5 + "\n"
	}
	return ret, nil
}
func (r *RemoteService) genMetaSignString(data *rmtt_msg.MetaData) string {
	var ret = ""
	ret = ret + utils.ProtoNumberToString(data.GetMsgType().Number()) + "\n"
	ret = ret + data.Nonce + "\n"
	ret = ret + utils.Int64ToString(data.GetTimestamp()) + "\n"
	if data.GetKey() != "" {
		ret = ret + data.GetKey() + "\n"
	}
	ret = ret + utils.ProtoNumberToString(data.GetSignatureMethod().Number()) + "\n"
	if data.GetBigDataMd5() != "" {
		ret = ret + data.BigDataMd5 + "\n"
	}
	ret = strings.TrimSuffix(ret, "\n")
	return ret
}
func (r *RemoteService) decryptData(enum rmtt_msg.EncryptionAlgorithmEnum, key, data []byte) ([]byte, error) {
	switch enum {
	case rmtt_msg.EncryptionAlgorithmEnum_NoEncryption:
		return data, nil
	case rmtt_msg.EncryptionAlgorithmEnum_AESGCM128Algorithm:

		ret, err := secure.DecryptData(secure.AESGCM128Algorithm, data, key)
		if err != nil {
			return nil, err
		}
		return ret, nil
	case rmtt_msg.EncryptionAlgorithmEnum_AESGCM256Algorithm:
		ret, err := secure.DecryptData(secure.AESGCM256Algorithm, data, key)
		if err != nil {
			return nil, err
		}
		return ret, nil
	case rmtt_msg.EncryptionAlgorithmEnum_AESGCM192Algorithm:
		ret, err := secure.DecryptData(secure.AESGCM192Algorithm, data, key)
		if err != nil {
			return nil, err
		}
		return ret, nil
	case rmtt_msg.EncryptionAlgorithmEnum_ChaCha20Poly1305Algorithm:
		ret, err := secure.DecryptData(secure.ChaCha20Poly1305Algorithm, data, key)
		if err != nil {
			return nil, err
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unknown encryption algorithm")
	}
}
func RmttSignAlgorithmToSignAlgorithm(enum rmtt_msg.SignatureMethod) secure.SignAlgorithm {
	switch enum {
	case rmtt_msg.SignatureMethod_NoSignature:
		return secure.None
	case rmtt_msg.SignatureMethod_HmacSHA256:
		return secure.HMAC_SHA256
	default:
		return secure.UNKNOWN
	}
}
func (r *RemoteService) ProtoHandler(data, sessionId []byte) {
	if r.client == nil {
		return
	}

	var msg rmtt_msg.RmttMsg
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		logger.Warn(err)
		r.sendResp("", "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
		return
	}

	switch msg.GetMetaData().GetMsgType() {
	case rmtt_msg.RmttMsgType_UnlockPcMsgType:
		{
			logger.Debug("receive unlock msg")
			msgData := msg.GetData()
			if msgData == nil {
				logger.Warn("unlock msg data is nil")
				return
			}
			var unlockMsg rmtt_msg.LoginToPCRequestMsg
			err := msgData.UnmarshalTo(&unlockMsg)
			if err != nil {
				logger.Warn(err)
				//todo
				r.sendResp("", "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}

			key, err := secure.DecodeBase58Key(r.credential.SecurityKey)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
			}
			userNameBase64 := base64.StdEncoding.EncodeToString(unlockMsg.Username)
			passwordBase64 := base64.StdEncoding.EncodeToString(unlockMsg.Password)
			reqSignString := r.genMetaSignString(msg.GetMetaData()) + "\n" + userNameBase64 + "\n" + passwordBase64
			signatureRet := secure.ValidateHMAC(reqSignString, r.credential.AccessSecret, msg.GetMetaData().GetSignature(), RmttSignAlgorithmToSignAlgorithm(msg.GetMetaData().GetSignatureMethod()))
			if !signatureRet {
				logger.Warn("signature check failed")
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
			}
			username, err := r.decryptData(unlockMsg.EncryptionMethod, key, unlockMsg.Username)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			password, err := r.decryptData(unlockMsg.EncryptionMethod, key, unlockMsg.Password)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			ex := r.un.UnlockPc(string(username), string(password))
			if ex != nil || exception.ErrSuccess.NotEqual(ex) {
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, ex.Code, ex.Msg, nil)
				return
			}
			r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrSuccess.Code, exception.ErrSuccess.Msg, nil)
			return

		}
	case rmtt_msg.RmttMsgType_LockScreenMsgType:
		{
			logger.Debug("receive lock screen msg")
			msgData := msg.GetData()
			if msgData == nil {
				logger.Warn("lock screen msg data is nil")
				return
			}
			var lockMsg rmtt_msg.LockScreenRequestMsg
			err := msgData.UnmarshalTo(&lockMsg)
			if err != nil {
				//todo
				logger.Warn(err)
			}
			r.co.LockWindows(false)
		}
	case rmtt_msg.RmttMsgType_ShutdownMsgType:
		{
			logger.Debug("receive shutdown msg")
			msgData := msg.GetData()
			if msgData == nil {
				logger.Warn("shutdown msg data is nil")
				return
			}
			var shutdownMsg rmtt_msg.ShutdownRequestMsg
			err := msgData.UnmarshalTo(&shutdownMsg)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			goroutine.RecoverGO(func() {
				time.Sleep(time.Duration(5) * time.Second)
				r.co.Shutdown(sys.ShutdownType(sys.E_FORCE_SHUTDOWN))
			})

			r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrSuccess.Code, exception.ErrSuccess.Msg, nil)
			return
		}
	case rmtt_msg.RmttMsgType_StandbyMsgType:
		{
			logger.Debug("receive standby msg")
			msgData := msg.GetData()
			if msgData == nil {
				logger.Warn("standby msg data is nil")
				return
			}
			var standbyMsg rmtt_msg.StandbyRequestMsg
			err := msgData.UnmarshalTo(&standbyMsg)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			goroutine.RecoverGO(func() {
				time.Sleep(time.Duration(5) * time.Second)
				r.co.Standby()
			})
			r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrSuccess.Code, exception.ErrSuccess.Msg, nil)
			return
		}
	case rmtt_msg.RmttMsgType_PowerOnMsgType:
		{
			logger.Debug("receive power on msg")
			msgData := msg.GetData()
			if msgData == nil {
				logger.Warn("power on msg data is nil")
				return
			}
			var powerOnMsg rmtt_msg.PowerOnRequestMsg
			err := msgData.UnmarshalTo(&powerOnMsg)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			macAddr, err := r.decryptData(powerOnMsg.EncryptionMethod, []byte(r.credential.SecurityKey), powerOnMsg.MacAddress)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
				return
			}
			err = r.co.PowerOnOtherDevices(macAddr)
			if err != nil {
				logger.Warn(err)
				r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
			}
			r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrSuccess.Code, exception.ErrSuccess.Msg, nil)

			return
		}

	default:
		logger.Warn("unknown msg type")
		r.sendResp(msg.MetaData.GetKey(), "", sessionId, exception.ErrUserParameterError.Code, exception.ErrUserParameterError.Msg, nil)
		return

	}

	return

}

func (r *RemoteService) RRFPMsgHandler(client RMTT.Client, msg RMTT.Message) {
	r.client = client
	dataSlice := msg.Payload()
	if len(dataSlice) == 0 {
		return
	}
	packet := &remote_schema.PayloadPacket{}
	err := packet.Unpack(dataSlice)
	if err != nil {
		logger.Warn(err)
		return
	}
	r.waitMapLock.RLock()
	if r.waitMap != nil {
		if ch, ok := r.waitMap[string(packet.SessionId)]; ok {
			r.waitMapLock.RUnlock()
			select {
			case ch <- packet:
				r.waitMapLock.Lock()
				delete(r.waitMap, string(packet.SessionId))
				r.waitMapLock.Unlock()
				close(ch)
			case <-r.remoteCtx.Done():
				r.waitMapLock.Lock()
				delete(r.waitMap, string(packet.SessionId))
				r.waitMapLock.Unlock()
				close(ch)
			}
		} else {
			r.waitMapLock.RUnlock()
		}
	} else {
		r.waitMapLock.RUnlock()
	}

	switch packet.DataType {
	case remote_schema.ProtoBuf:
		r.ProtoHandler(packet.Data, packet.SessionId)
	default:
		return
	}

}

//	func (r *RemoteService) PushTextRet(client RMTT.Client, ret *exception.Exception, requestId []byte) {
//		if len(requestId) > 0xff {
//			return
//		}
//		requestIdLen := uint8(len(requestId))
//		buffer := new(bytes.Buffer)
//		err := binary.Write(buffer, binary.BigEndian, ret.Code)
//		if err != nil {
//			logger.Warn(err)
//			return
//		}
//
//		payload := buffer.Bytes()
//		packet := &remote_schema.PayloadPacket{DataType: remote_schema.Text,
//			Data: payload, SessionIdLen: requestIdLen, SessionId: requestId}
//		if client != nil {
//
//			client.Push(packet)
//		}
//	}
func (r *RemoteService) buildProtobufResp(requestId string, code int, msg string, data []byte) *rmtt_msg.RmttResponseMsg {

	resp := &rmtt_msg.RmttResponseMsg{
		RequestId: requestId,
		Code:      int32(code),
		Msg:       msg,
		BigData:   data,
	}
	return resp
}
func (r *RemoteService) sendResp(accessKey, requestId string, sessionId []byte, code int, msg string, data []byte) error {
	resp := r.buildProtobufResp(requestId, code, msg, data)
	anyData, err := anypb.New(resp)
	if err != nil {
		return err
	}
	meta := rmtt_msg.MetaData{
		MsgType:         rmtt_msg.RmttMsgType_RfuResponseMsgType,
		Timestamp:       time.Now().UnixMilli(),
		Nonce:           uuid.New().String(),
		Key:             accessKey,
		SignatureMethod: rmtt_msg.SignatureMethod_NoSignature,
	}
	rmttMsg := rmtt_msg.RmttMsg{
		MetaData: &meta,
		Data:     anyData,
	}

	bytes, err := proto.Marshal(&rmttMsg)
	if err != nil {
		return err
	}
	err = r.SendData(remote_schema.ProtoBuf, sessionId, bytes)
	if err != nil {
		return err
	}
	return nil

}
func (r *RemoteService) SendData(dataType remote_schema.PacketType, sessionId, data []byte) error {
	if len(sessionId) > 0xff {
		return fmt.Errorf("sessionId too long")
	}
	requestIdLen := uint8(len(sessionId))

	packet := &remote_schema.PayloadPacket{SessionIdLen: requestIdLen, SessionId: sessionId, DataType: dataType, Data: data}

	ret, err := packet.Pack()
	if err != nil {

		return err
	}
	if r.client != nil {
		r.client.Push(ret)
	} else {
		return fmt.Errorf("no client")
	}
	return nil
}
func (r *RemoteService) OnProtoMessage(sessionId string) *rmtt_msg.RmttMsg {

	r.waitMapLock.Lock()
	if r.waitMap == nil {
		r.waitMap = make(map[string]chan *remote_schema.PayloadPacket)
	}
	ch := make(chan *remote_schema.PayloadPacket)
	r.waitMap[sessionId] = ch
	r.waitMapLock.Unlock()
	select {
	case <-r.remoteCtx.Done():
		return nil
	case packet := <-ch:
		if packet == nil {
			return nil
		}
		var msg rmtt_msg.RmttMsg
		err := proto.Unmarshal(packet.Data, &msg)
		if err != nil {
			logger.Warn(err)
			return nil
		}
		return &msg

	}

}

//func (r *RemoteService) sendData(client RMTT.Client, topic []byte, dataType remote_schema.PacketType, data []byte) {
//	if len(topic) > 0xff {
//		return
//	}
//	requestIdLen := uint8(len(topic))
//
//	packet := &remote_schema.PayloadPacket{SessionIdLen: requestIdLen, SessionId: topic, DataType: dataType, Data: data}
//
//	ret, err := packet.Pack()
//	if err != nil {
//		logger.Error(err)
//		return
//	}
//	if client != nil {
//		client.Push(ret)
//	}
//}

type rfuServerResponse[T any] struct {
	RequestId string `json:"request_id"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      T      `json:"data"`
}

func (r *RemoteService) refreshToken() (*remote_schema.DeviceRefreshTokenResponse, error) {
	if r.apiServerUrl == "" {
		return nil, fmt.Errorf("apiServerUrl is empty")
	}
	url := r.apiServerUrl + "/api/v1/devices/refresh-token"
	client := utils.NewClient()
	var req = remote_schema.DeviceRefreshTokenRequest{DeviceId: r.clientId}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	resp, err := client.PostWithSign(url, r.accessKey, r.accessSecret, secure.HMAC_SHA256, map[string]string{
		"Content-Type": "application/json",
	}, body)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}
	if resp.StatusCode != http.StatusOK {
		var result rfuServerResponse[interface{}]
		err := json.Unmarshal(bodyData, &result)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected response code: %d ,msg: %s", result.Code, result.Msg)
	}
	var result rfuServerResponse[remote_schema.DeviceRefreshTokenResponse]
	err = json.Unmarshal(bodyData, &result)
	if err != nil {
		return nil, err
	}

	return &result.Data, nil

}
func (r *RemoteService) GetMsgServerUrl() (*[]remote_schema.RmttMsgUrls, error) {
	if r.apiServerUrl == "" {
		return nil, fmt.Errorf("apiServerUrl is empty")
	}
	url := r.apiServerUrl + "/api/v1/devices/msg-server-urls"
	client := utils.NewClient()
	resp, err := client.GetWithSign(url, r.accessKey, r.accessSecret, secure.HMAC_SHA256, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}
	var result rfuServerResponse[[]remote_schema.RmttMsgUrls]
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.Code != exception.ErrSuccess.Code {
		return nil, fmt.Errorf("unexpected response code: %s", result.Code)
	}
	return &result.Data, nil
}
func (r *RemoteService) RegisterDevice() (string, error) {
	if r.apiServerUrl == "" {
		return "", fmt.Errorf("apiServerUrl is empty")
	}

	client := utils.NewClient()
	url := r.apiServerUrl + "/api/v1/devices/register"

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	var req = remote_schema.RegisterDeviceRequest{
		DeviceName: hostname,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := client.PostWithSign(url, r.accessKey, r.accessSecret, secure.HMAC_SHA256, map[string]string{
		"Content-Type": "application/json",
	}, body)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var result rfuServerResponse[remote_schema.RegisterDeviceResponse]
	err = json.Unmarshal(bodyData, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	if result.Code != exception.ErrSuccess.Code {
		return "", fmt.Errorf("unexpected response code: %s", result.Code)
	}
	clientId := result.Data.DeviceId
	if clientId == "" {
		return "", fmt.Errorf("device id is empty")
	}
	return clientId, nil
}

func (r *RemoteService) GetConfig() (*remote_schema.RemoteConfigResponse, error) {
	var config entity.RemoteConfig
	if err := r.db.First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config = entity.RemoteConfig{
				Enable:                false,
				DefaultRemoteServerId: 0,
			}
			r.db.Create(&config)
			return &remote_schema.RemoteConfigResponse{
				Enable:      false,
				ApiServerId: 0,
			}, nil
		}
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	return &remote_schema.RemoteConfigResponse{
		Enable:      config.Enable,
		ApiServerId: uint32(config.DefaultRemoteServerId),
	}, nil

}
func (r *RemoteService) UpdateRemoteConfig(data *remote_schema.RemoteConfigRequest) error {

	config := entity.RemoteConfig{}
	if err := r.db.First(&config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return fmt.Errorf("failed to find database: %v", err)
	}

	if config.Enable != data.Enable {
		defer func() {
			r.RestartService()
		}()
	}
	config.Enable = data.Enable
	if uint(data.ApiServerId) != 0 {
		config.DefaultRemoteServerId = uint(data.ApiServerId)
	}
	if err := r.db.Save(&config).Error; err != nil {
		logger.Errorf("failed to save database: %v", err)
		return fmt.Errorf("failed to save database: %v", err)
	}
	return nil

}
func (r *RemoteService) PatchRemoteConfig(content map[string]interface{}) error {

	var config entity.RemoteConfig
	if err := r.db.First(&config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return fmt.Errorf("failed to find database: %v", err)
	}

	if err := r.db.Model(&config).Updates(content).Error; err != nil {
		return err
	}

	return nil

}
func (r *RemoteService) GetCredential() (*remote_schema.CredentialResponse, error) {
	var credential entity.Credential
	if err := r.db.First(&credential).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	if credential.AccessKey == "" || credential.AccessSecret == "" || credential.SecurityKey == "" {
		accessKey, accessSecret, key, err := r.genCredentials()
		if err != nil {
			return nil, err
		}
		credential = entity.Credential{
			AccessKey:    accessKey,
			AccessSecret: accessSecret,
			SecurityKey:  key,
		}
		if err := r.db.Save(&credential).Error; err != nil {
			logger.Errorf("failed to save database: %v", err)
			return nil, fmt.Errorf("failed to save database: %v", err)
		}
		return &remote_schema.CredentialResponse{
			AccessKey:    accessKey,
			AccessSecret: accessSecret,
			SecurityKey:  key,
		}, nil
	}
	r.credential = credential
	return &remote_schema.CredentialResponse{
		AccessKey:    credential.AccessKey,
		AccessSecret: credential.AccessSecret,
		SecurityKey:  credential.SecurityKey,
	}, nil

}
func (r *RemoteService) RefreshCredential() (*remote_schema.CredentialResponse, error) {

	r.db.Exec("delete from  credentials")

	accessKey, accessSecret, key, err := r.genCredentials()
	if err != nil {
		return nil, err
	}
	credential := entity.Credential{
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
		SecurityKey:  key,
	}
	if err := r.db.Save(&credential).Error; err != nil {
		logger.Errorf("failed to save database: %v", err)
		return nil, fmt.Errorf("failed to save database: %v", err)
	}
	r.credential = credential
	return &remote_schema.CredentialResponse{
		AccessKey:    credential.AccessKey,
		AccessSecret: credential.AccessSecret,
		SecurityKey:  credential.SecurityKey,
	}, nil

}
func (r *RemoteService) GetRemoteApiServerConfig(id int) (*remote_schema.RemoteApiServerConfigResponse, error) {

	var server entity.RemoteServer
	if err := r.db.First(&server, id).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}

	var msgServers []entity.RemoteRmttServers
	if err := r.db.Where("remote_server_id = ?", server.ID).Find(&msgServers).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	msgServerUrls := make([]remote_schema.MsgServerUrl, 0)
	for _, msgServer := range msgServers {
		msgServerUrls = append(msgServerUrls, remote_schema.MsgServerUrl{
			Id:           uint32(msgServer.ID),
			MsgServerUrl: msgServer.MsgServerUrl,
			Weight:       msgServer.Weight,
			Enable:       msgServer.Enable,
		})
	}
	return &remote_schema.RemoteApiServerConfigResponse{
		Id:                   uint32(server.ID),
		ApiServerUrl:         server.ApiServerUrl,
		AccessKey:            server.AccessKey,
		AccessSecret:         server.AccessSecret,
		Token:                server.Token,
		TokenExpiresAt:       server.TokenExpiresAt,
		ClientId:             server.ClientId,
		EnableSignatureCheck: server.EnableSignatureCheck,
		MsgServerUrls:        msgServerUrls,
	}, nil
}
func (r *RemoteService) UpdateRemoteApiServerConfig(id int, request *remote_schema.RemoteApiServerConfigRequest) (*remote_schema.RemoteApiServerConfigResponse, error) {
	var server entity.RemoteServer
	if err := r.db.First(&server, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apiServerUrl := request.ApiServerUrl
			apiServerUrl = strings.TrimSpace(apiServerUrl)
			apiServerUrl = strings.TrimSuffix(apiServerUrl, "/")
			server = entity.RemoteServer{
				ApiServerUrl:         apiServerUrl,
				AccessKey:            request.AccessKey,
				AccessSecret:         request.AccessSecret,
				EnableSignatureCheck: request.EnableSignatureCheck,
			}
			if err := r.db.Save(&server).Error; err != nil {
				logger.Errorf("failed to save database: %v", err)
				return nil, fmt.Errorf("failed to save database: %v", err)
			}
			return &remote_schema.RemoteApiServerConfigResponse{
				Id:                   uint32(server.ID),
				ApiServerUrl:         server.ApiServerUrl,
				AccessKey:            server.AccessKey,
				AccessSecret:         server.AccessSecret,
				EnableSignatureCheck: server.EnableSignatureCheck,
			}, nil
		}
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	if request.ApiServerUrl != "" {
		server.ApiServerUrl = request.ApiServerUrl
	}
	if request.AccessKey != "" {
		server.AccessKey = request.AccessKey
	}
	if request.AccessSecret != "" {
		server.AccessSecret = request.AccessSecret
	}

	server.EnableSignatureCheck = request.EnableSignatureCheck
	if err := r.db.Save(&server).Error; err != nil {
		logger.Errorf("failed to save database: %v", err)
		return nil, fmt.Errorf("failed to save database: %v", err)
	}
	var config entity.RemoteConfig
	if err := r.db.First(&config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return nil, fmt.Errorf("failed to find database: %v", err)
	}
	config.DefaultRemoteServerId = server.ID
	if err := r.db.Save(&config).Error; err != nil {
		logger.Errorf("failed to save database: %v", err)
		return nil, fmt.Errorf("failed to save database: %v", err)
	}
	return &remote_schema.RemoteApiServerConfigResponse{
		Id:                   uint32(server.ID),
		ApiServerUrl:         server.ApiServerUrl,
		AccessKey:            server.AccessKey,
		AccessSecret:         server.AccessSecret,
		EnableSignatureCheck: server.EnableSignatureCheck,
	}, nil

}

func (r *RemoteService) RestartService() error {
	err := r.StopService()
	if err != nil {
		return fmt.Errorf("stop remote service error: %v", err)
	}
	err = r.StartService()
	if err != nil {
		return fmt.Errorf("start remote service error: %v", err)
	}
	logger.Debug("restarting service")
	return nil
}

func (r *RemoteService) loadConfig() error {
	config := entity.RemoteConfig{}
	if err := r.db.First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.enable = false
			return nil
		}
		return err
	}
	r.enable = config.Enable
	if !config.Enable {
		return fmt.Errorf("remote service is disabled")
	}

	serverId := config.DefaultRemoteServerId
	server := entity.RemoteServer{}
	if err := r.db.First(&server, serverId).Error; err != nil {
		return fmt.Errorf("failed to find remote server: %v", err)
	}
	r.defaultApiServerId = serverId
	r.apiServerUrl = server.ApiServerUrl
	r.accessKey = server.AccessKey
	r.accessSecret = server.AccessSecret
	if server.ClientId == "" {
		clientId, e := r.RegisterDevice()
		if e != nil {
			return e
		}
		server.ClientId = clientId
	}
	r.clientId = server.ClientId
	expireTime := time.UnixMilli(server.TokenExpiresAt)
	currentTime := time.Now()
	if server.Token == "" || currentTime.After(expireTime) {
		resp, err := r.refreshToken()
		if err != nil {
			return err
		}
		server.Token = resp.NewToken
		server.TokenExpiresAt = resp.ExpiresAt
	}
	r.db.Save(&server)
	r.accessToken = server.Token
	err := r.db.Model(entity.RemoteRmttServers{RemoteServerId: server.ID}).Limit(10).Find(&r.msgServers).Order("weight desc").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err := r.loadMsgServers(server.ID)
		if err != nil {
			return err
		}
		accessKey, accessSecret, key, err := r.genCredentials()
		if err != nil {
			return err
		}
		r.credential = entity.Credential{
			AccessKey:    strings.ToLower(accessKey),
			AccessSecret: strings.ToLower(accessSecret),
			SecurityKey:  key,
		}

	} else if err != nil {
		return err
	}
	if err := r.db.First(&r.credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			r.db.Save(&r.credential)
		} else {
			return err
		}
	}
	return nil
}
func (r *RemoteService) loadMsgServers(apiServerId uint) error {
	ret, e := r.GetMsgServerUrl()
	if e != nil {
		return e
	}
	msgServers := *ret
	sort.Slice(msgServers, func(i, j int) bool {
		return msgServers[i].Weight > msgServers[j].Weight
	})
	r.msgServers = make([]entity.RemoteRmttServers, 0)
	for _, msgServer := range msgServers {
		server := entity.RemoteRmttServers{
			MsgServerUrl:   msgServer.Url,
			RemoteServerId: apiServerId,
			Weight:         msgServer.Weight,
			Enable:         true,
		}
		r.msgServers = append(r.msgServers, server)
		r.db.Save(&server)
	}
	return nil
}
func (r *RemoteService) genCredentials() (accessKey, accessSecret, securityKey string, err error) {
	accessKey, err = secure.GenerateRandomBase58Key(8)
	if err != nil {
		return "", "", "", err
	}
	accessSecret, err = secure.GenerateRandomBase58Key(16)
	if err != nil {
		return "", "", "", err
	}
	securityKey, err = secure.GenerateRandomBase58Key(35)
	if err != nil {
		return "", "", "", err
	}

	return accessKey, accessSecret, securityKey, nil
}
func (r *RemoteService) TestServerDelay() (int64, error) {
	msg := rmtt_msg.PingMsg{}
	data, err := anypb.New(&msg)
	if err != nil {
		return -1, err
	}
	sessionId := uuid.New().String()
	pingMsg := rmtt_msg.RmttMsg{
		MetaData: &rmtt_msg.MetaData{
			MsgType: rmtt_msg.RmttMsgType_PingMsgType,
		},
		Data: data,
	}

	buf, err := proto.Marshal(&pingMsg)
	if err != nil {
		return -1, err
	}
	now := time.Now().UnixNano()
	err = r.SendData(remote_schema.ProtoBuf, []byte(sessionId), buf)
	if err != nil {
		return -1, err
	}
	logger.Debug("send ping msg")
	r.OnProtoMessage(sessionId)
	logger.Debug("receive pong msg")
	delay := time.Now().UnixNano() - now
	return delay, nil
}

func (r *RemoteService) StartService() error {
	r.remoteCtx, r.remoteCancel = context.WithCancel(r.ctx)
	var err error
	defer func() {
		r.lastError = err
	}()

	logger.Debug("starting service")
	err = r.loadConfig()
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return err
	}
	if !r.enable {
		return err
	}
	err = r.loadMsgServers(r.defaultApiServerId)
	if err != nil {
		logger.Errorf("failed to load msg servers: %v", err)
		return err
	}
	RMTT.DEBUG = logger.GetLogger()
	RMTT.ERROR = logger.GetLogger()
	RMTT.INFO = logger.GetLogger()
	RMTT.WARN = logger.GetLogger()

	opts := RMTT.NewClientOptions()
	for _, server := range r.msgServers {
		opts.AddServer(server.MsgServerUrl)
	}
	//}
	opts.SetToken(r.accessToken)
	opts.AutoReconnect = true
	logger.Debug("your client id is ", r.clientId)
	client := RMTT.NewClient(opts)
	client.AddPayloadHandlerLast(r.RRFPMsgHandler)
	r.client = client
	go func() {
		token := client.Connect()
		token.Wait()
		err := token.Error()
		if err != nil {
			logger.Warnf("connect fail%v", token.Error())
			if errors.Is(err, RMTT.RefusedNotAuthorisedErr) || errors.Is(err, RMTT.ProtocolViolationErr) {
				r.lastError = err
				return
			}
			select {
			case <-r.remoteCtx.Done():
				client.Disconnect(0)
				r.StopService()
				return
			}

			return
		}

	}()

	select {
	case <-r.remoteCtx.Done():
		client.Disconnect(0)
		r.StopService()
		return nil
	}

}
func (r *RemoteService) StopService() error {
	r.remoteCancel()
	return nil
}
