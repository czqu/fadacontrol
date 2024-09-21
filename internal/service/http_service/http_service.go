package http_service

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema/http_schema"
	"fadacontrol/pkg/secure"
	"fmt"

	"gorm.io/gorm"
)

type HttpService struct {
	_db *gorm.DB
}

func NewHttpService(_db *gorm.DB) *HttpService {
	return &HttpService{_db: _db}
}

const HttpServiceApi = "HTTP_SERVICE_API"
const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func (s HttpService) GetHttpConfig(serviceName string) (interface{}, error) {

	if serviceName == HttpServiceApi {
		var httpConfig entity.HttpConfig
		if err := s._db.First(&httpConfig).Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return nil, fmt.Errorf("failed to find database: %v", err)
		}
		return &http_schema.HttpConfigResponse{
			Enable: httpConfig.Enable,
			Host:   httpConfig.Host,
			Port:   httpConfig.Port,
		}, nil
	}

	if serviceName == HttpsServiceApi {
		var httpsConfig entity.HttpConfig
		if err := s._db.First(&httpsConfig).Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return nil, fmt.Errorf("failed to find database: %v", err)
		}
		return &http_schema.HttpsConfigResponse{
			Enable: httpsConfig.Enable,
			Host:   httpsConfig.Host,
			Port:   httpsConfig.Port,
			Cer:    httpsConfig.Cer,
			Key:    httpsConfig.Key,
		}, nil
	}

	//todo
	return nil, fmt.Errorf("failed to find database")
}

func (s HttpService) UpdateHttpConfig(h interface{}, serviceName string) error {
	if request, ok := h.(http_schema.HttpConfigRequest); ok {
		var httpConfig entity.HttpConfig
		if err := s._db.First(&httpConfig).Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}
		httpConfig.Enable = request.Enable
		httpConfig.Host = request.Host
		httpConfig.Port = request.Port
		err := s._db.Save(&httpConfig).Error
		if err != nil {
			return fmt.Errorf("failed to save http config: %v", err)
		}
	}
	if request, ok := h.(http_schema.HttpsConfigRequest); ok {
		_, err := secure.LoadBaseX509KeyPair(request.Cer, request.Key)
		if err != nil {
			logger.Error(err)
			//todo
			return fmt.Errorf("failed to load https config: %v", err)
		}

		var httpsConfig entity.HttpConfig
		if err := s._db.First(&httpsConfig).Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}

		httpsConfig.Enable = request.Enable
		httpsConfig.Host = request.Host
		httpsConfig.Port = request.Port
		httpsConfig.Key = request.Key
		httpsConfig.Cer = request.Cer

		err = s._db.Save(&httpsConfig).Error
		if err != nil {
			return fmt.Errorf("failed to save http config: %v", err)
		}
	}

	//todo
	return fmt.Errorf("unknow")

}

func (s HttpService) PatchHttpConfig(data map[string]interface{}, serviceName string) error {
	if serviceName == HttpServiceApi {

		var config entity.HttpConfig
		if err := s._db.First(&config).Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}
		if err := s._db.Model(&config).Updates(data).Error; err != nil {

			logger.Errorf("failed to patch http config: %v", err)
			return fmt.Errorf("failed to patch http config: ")
		}
		return nil

	}
	if serviceName == HttpsServiceApi {
		var config entity.HttpConfig
		if err := s._db.First(&config).Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}
		if err := s._db.Model(&config).Updates(data).Error; err != nil {
			return err
		}
		return nil
	}

	//todo
	return fmt.Errorf("unknow")
}
