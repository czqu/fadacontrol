package http_service

import (
	"crypto/tls"
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/goroutine"
	"golang.org/x/net/context"
	"time"

	"fadacontrol/internal/entity"
	"fadacontrol/internal/router"
	"fadacontrol/internal/schema/http_schema"
	"fadacontrol/pkg/secure"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"gorm.io/gorm"
	"net/http"
)

type HttpService struct {
	signalChanMap        map[string]chan interface{}
	_db                  *gorm.DB
	_conf                *conf.Conf
	adminRouter          router.FadaControlRouter
	commonRouter         router.FadaControlRouter
	restartAllServerFunc func() error
}

func NewHttpService(_db *gorm.DB, _conf *conf.Conf) *HttpService {
	return &HttpService{_db: _db, _conf: _conf, signalChanMap: make(map[string]chan interface{})}
}

const HttpServiceApi = "HTTP_SERVICE_API"
const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func (s *HttpService) GetHttpConfig(serviceName string) (interface{}, error) {

	if serviceName == HttpServiceApi {
		var httpConfig entity.HttpConfig
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).First(&httpConfig).Error; err != nil {
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
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).First(&httpsConfig).Error; err != nil {
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

	return nil, exception.ErrUserParameterError
}

func (s *HttpService) UpdateHttpConfig(h interface{}, serviceName string) error {
	if request, ok := h.(http_schema.HttpConfigRequest); ok {
		if serviceName != HttpServiceApi {
			return exception.ErrUserParameterError
		}
		var httpConfig entity.HttpConfig
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).First(&httpConfig).Error; err != nil {
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
		if serviceName != HttpServiceApi {
			return exception.ErrUserParameterError
		}
		_, err := secure.LoadBaseX509KeyPair(request.Cer, request.Key)
		if err != nil {
			logger.Error(err)
			return fmt.Errorf("failed to load https config: %v", err)
		}

		var httpsConfig entity.HttpConfig
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).First(&httpsConfig).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}

		httpsConfig.Enable = request.Enable
		httpsConfig.Host = request.Host
		httpsConfig.Port = request.Port
		httpsConfig.Key = request.Key
		httpsConfig.Cer = request.Cer
		httpsConfig.EnableHttp3 = request.EnableHttp3

		err = s._db.Save(&httpsConfig).Error
		if err != nil {
			return fmt.Errorf("failed to save http config: %v", err)
		}
	}

	return exception.ErrUserParameterError

}

func (s *HttpService) PatchHttpConfig(data map[string]interface{}, serviceName string) error {
	if serviceName == HttpServiceApi {

		var config entity.HttpConfig
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpServiceApi}).First(&config).Error; err != nil {
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
		if err := s._db.Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).First(&config).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}
		if err := s._db.Model(&config).Updates(data).Error; err != nil {
			return err
		}
		return nil
	}

	return exception.ErrUserParameterError
}

func (s *HttpService) StartServer(r router.FadaControlRouter, serviceName string) error {
	enableQuic := false

	if s._conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if serviceName == HttpServiceAdmin {
		s.adminRouter = r
	}

	if serviceName == HttpServiceApi || serviceName == HttpsServiceApi {
		s.commonRouter = r
	}

	if serviceName == HttpServiceApi || serviceName == HttpServiceAdmin || serviceName == HttpsServiceApi {
		var config entity.HttpConfig
		if err := s._db.Where(&entity.HttpConfig{ServiceName: serviceName}).First(&config).Error; err != nil {
			logger.Errorf("failed to find database: %v", err)
			return fmt.Errorf("failed to find database: %v", err)
		}
		if config.Enable == false {
			return nil
		}
		logger.Infof("Starting HTTP server on %s:%d ", config.Host, config.Port)
		r.Register()
		cert := tls.Certificate{}
		if serviceName == HttpsServiceApi {
			var err error
			cert, err = secure.LoadBaseX509KeyPair(config.Cer, config.Key)
			if err != nil {
				logger.Error(err)
				_cert, _key, err := secure.GenerateX509Cert()
				if err != nil {
					logger.Error("gen key err", err)
					return err
				}
				cert, err = secure.LoadX509KeyPairFromMemory(_cert, _key)
				if err != nil {
					logger.Error(err)
					return err
				}
				return err
			}
			enableQuic = config.EnableHttp3
		}

		_router := r.GetRouter()
		sign := make(chan interface{})
		s.signalChanMap[config.ServiceName] = sign
		goroutine.RecoverGO(func() {
			startHttpServer(config.Host, config.Port, cert, _router, enableQuic, sign)
		})
		return nil
	}

	return nil
}

func startHttpServer(host string, port int, cert tls.Certificate, router *gin.Engine, enableEnableHttp3 bool, sign chan interface{}) {
	var srv *http.Server
	var http3Server *http3.Server
	tlsFlag := false

	if cert.Certificate == nil || len(cert.Certificate) == 0 || cert.PrivateKey == nil {
		tlsFlag = false
		logger.Info("start no secure http at ", port)
		srv = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: router,
		}

	} else {
		tlsFlag = true
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			MinVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		}

		if enableEnableHttp3 {
			conf.Http3Enabled = true
			conf.Http3Port = port
			http3Server = &http3.Server{
				Addr:       fmt.Sprintf(":%d", port),
				Handler:    router,
				TLSConfig:  tlsConfig,
				QUICConfig: &quic.Config{},
			}
			logger.Info("start http3 server at ", port)

			goroutine.RecoverGO(func() {
				err := http3Server.ListenAndServe()
				if err != nil {
					logger.Error(err)
				}
			})

		}

		srv = &http.Server{
			Addr:      fmt.Sprintf(":%d", port),
			Handler:   router,
			TLSConfig: tlsConfig,
		}
		logger.Info("start secure server at ", port)

	}
	goroutine.RecoverGO(func() {
		<-sign
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("server Shutdown: %s", err)
			close(sign)
		}
		if http3Server == nil {
			return
		}
		err := http3Server.Close()
		if err != nil {
			logger.Errorf("http3 server Close Err: %s", err)
			return
		}
	})
	var err error
	if tlsFlag || enableEnableHttp3 {
		err = srv.ListenAndServeTLS("", "")

	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		logger.Warnf("server errors: %s", err)
	}

}
func (s *HttpService) StopServer(serviceName string) error {

	if sign, ok := s.signalChanMap[serviceName]; ok {
		sign <- 0
		return nil
	}

	return errors.New("not found http service,name: " + serviceName)
}

func (s *HttpService) StopAllServer() error {

	for name, sign := range s.signalChanMap {
		logger.Info("stopping http service: " + name)
		sign <- 0
	}
	return nil
}
