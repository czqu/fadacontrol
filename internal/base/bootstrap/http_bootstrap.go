package bootstrap

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/router"
	"fadacontrol/pkg/secure"
	"gorm.io/gorm"
	"net/http"

	"fmt"
	"github.com/gin-gonic/gin"
)

type HttpBootstrap struct {
	configs       []entity.HttpConfig
	db            *gorm.DB
	signalChanMap map[string]chan interface{}
	adr           *router.AdminRouter
	common        *router.CommonRouter
	_conf         *conf.Conf
}

func NewHttpBootstrap(_conf *conf.Conf, db *gorm.DB, adr *router.AdminRouter, common *router.CommonRouter) *HttpBootstrap {
	return &HttpBootstrap{_conf: _conf, db: db, adr: adr, common: common, signalChanMap: make(map[string]chan interface{})}
}

func (s *HttpBootstrap) Start() error {
	s.CreateConfig()
	return s.StartServer()

}
func (s *HttpBootstrap) Stop() error {
	return s.StopAllServer()

}

func (s *HttpBootstrap) refreshData() {

	if err := s.db.Find(&s.configs).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return
	}
	s.signalChanMap = make(map[string]chan interface{})
}

func (s *HttpBootstrap) StartServer() error {
	s.refreshData()

	if s._conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	for _, config := range s.configs {
		if config.Enable == false {
			continue
		}
		if config.ServiceName == HttpServiceApi {
			logger.Infof("Starting HTTP server on %s:%d ", config.Host, config.Port)
			s.common.Register()
			router := s.common.GetRouter()
			sign := make(chan interface{})
			s.signalChanMap[config.ServiceName] = sign
			go startHttpServer(config.Host, config.Port, tls.Certificate{}, router, sign)
		}
		if config.ServiceName == HttpsServiceApi {
			logger.Infof("Starting HTTPS server on %s:%d ", config.Host, config.Port)
			cert, err := secure.LoadBaseX509KeyPair(config.Cer, config.Key)
			if err != nil {
				logger.Error(err)
				continue
			}
			s.common.Register()
			router := s.common.GetRouter()
			sign := make(chan interface{})
			s.signalChanMap[config.ServiceName] = sign
			go startHttpServer(config.Host, config.Port, cert, router, sign)

		}
		if config.ServiceName == HttpServiceAdmin {
			logger.Infof("Starting HTTP server on %s:%d ", config.Host, config.Port)
			s.adr.Register()
			router := s.adr.GetRouter()
			sign := make(chan interface{})
			s.signalChanMap[config.ServiceName] = sign
			go startHttpServer(config.Host, config.Port, tls.Certificate{}, router, sign)
		}
	}
	return nil
}

func startHttpServer(host string, port int, cert tls.Certificate, router *gin.Engine, sign chan interface{}) {
	var srv *http.Server

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
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		}
		srv = &http.Server{
			Addr:      fmt.Sprintf(":%d", port),
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		logger.Info("start secure server at ", port)
	}
	go func() {
		<-sign
		if err := srv.Shutdown(nil); err != nil {
			logger.Error("server Shutdown: %s", err)
			close(sign)
		}
	}()
	var err error
	if tlsFlag {
		err = srv.ListenAndServeTLS("", "")

	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		logger.Error("server errors: %s", err)
	}

}
func (s *HttpBootstrap) StopServer(serviceName string) error {

	if sign, ok := s.signalChanMap[serviceName]; ok {
		sign <- 0
		return nil
	}

	return errors.New("not found http service,name: " + serviceName)
}
func (s *HttpBootstrap) StopAllServer() error {

	for name, sign := range s.signalChanMap {
		logger.Info("stopping http service: " + name)
		sign <- 0
	}
	return nil
}

const HttpServiceApi = "HTTP_SERVICE_API"
const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func (s *HttpBootstrap) CreateConfig() {

	err := s.db.AutoMigrate(&entity.HttpConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")
		return
	}
	var httpCount int64
	s.db.Model(&entity.HttpConfig{}).Count(&httpCount)
	if httpCount == 0 {
		cert, key, err := secure.GenerateX509Cert()
		if err != nil {
			logger.Errorf("failed to generate x509 cert: %v", err)
			return
		}
		strCert := base64.StdEncoding.EncodeToString(cert)
		strKey := base64.StdEncoding.EncodeToString(key)
		httpConfig := entity.HttpConfig{
			ServiceName: HttpServiceApi,
			Enable:      true,
			Host:        "0.0.0.0",
			Port:        2092,
			Cer:         "",
			Key:         "",
		}
		s.db.Create(&httpConfig)

		httpsConfig := entity.HttpConfig{
			ServiceName: HttpsServiceApi,
			Enable:      true,
			Host:        "0.0.0.0",
			Port:        2091,
			Cer:         strCert,
			Key:         strKey,
		}
		s.db.Create(&httpsConfig)

		httpAdminConfig := entity.HttpConfig{
			ServiceName: HttpServiceAdmin,
			Enable:      true,
			Host:        "127.0.0.1",
			Port:        2093,
			Cer:         "",
			Key:         "",
		}
		s.db.Create(&httpAdminConfig)
	}
}
