package bootstrap

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/pkg/secure"

	"fadacontrol/internal/entity"

	"gorm.io/gorm"
	"net"
	"strconv"
	"time"
)

type LegacyBootstrap struct {
	db            *gorm.DB
	signalChanMap map[string]chan interface{}
	configs       []entity.SocketServerConfig
	u             *unlock.UnLockService
	l             *control_pc.LegacyControlService
}

func NewLegacyBootstrap(u *unlock.UnLockService, db *gorm.DB, l *control_pc.LegacyControlService) *LegacyBootstrap {
	return &LegacyBootstrap{db: db, signalChanMap: make(map[string]chan interface{}), l: l, u: u}

}

func (s *LegacyBootstrap) Start() error {
	s.CreateConfig()
	return s.StartServer()

}
func (s *LegacyBootstrap) Stop() error {
	return s.StopAllServer()
}

//func init() {
//	RegisterService(&LegacyBootstrap{})
//}

func (s *LegacyBootstrap) refreshData() {

	if err := s.db.Find(&s.configs).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
		return
	}
	s.signalChanMap = make(map[string]chan interface{})
}
func (s *LegacyBootstrap) StartServer() error {
	s.refreshData()

	for _, config := range s.configs {
		if config.Enable == false {
			continue
		}
		if config.ServiceName == UnlockService {
			logger.Infof("Starting Socket server on %s:%d", config.Host, config.Port)
			sign := make(chan interface{})
			s.signalChanMap[config.ServiceName] = sign
			cert, err := secure.LoadBaseX509KeyPair(config.Cer, config.Key)
			if err != nil {
				logger.Error("can not load cert:", err)
				continue
			}
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS13,
			}
			go s.startSocketServer(config.Host, config.Port, sign, tlsConfig, s.u.HandleUnlockConnection)
		}
		if config.ServiceName == ControlService {
			logger.Infof("Starting Socket server on %s:%d", config.Host, config.Port)
			sign := make(chan interface{})
			s.signalChanMap[config.ServiceName] = sign
			cert, err := secure.LoadBaseX509KeyPair(config.Cer, config.Key)
			if err != nil {
				logger.Error("can not load cert:", err)
				continue
			}
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS13,
			}
			go s.startSocketServer(config.Host, config.Port, sign, tlsConfig, s.l.HandleControlConnection)
		}
	}
	return nil
}

func (s *LegacyBootstrap) startSocketServer(host string, port int, sign chan interface{}, config *tls.Config, handler func(conn net.Conn)) {
	addr := host + ":" + strconv.Itoa(port)
	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		logger.Error("unable to listen on port ", port)
		return
	}
	defer listener.Close()

	go func() {
		<-sign
		listener.Close()
		logger.Info("socket close at", port)
	}()
	logger.Info("socket wait for connection...")
	for {

		conn, err := listener.Accept()
		if err != nil {
			logger.Info("unable to accept connection", err)

			return
		}
		go handler(conn)
		time.Sleep(500 * time.Millisecond)
	}
}
func (s *LegacyBootstrap) StopServer(serviceName string) error {

	if sign, ok := s.signalChanMap[serviceName]; ok {
		sign <- 0
		return nil
	}

	return errors.New("not found http service,name: " + serviceName)

}
func (s *LegacyBootstrap) StopAllServer() error {

	for name, sign := range s.signalChanMap {
		logger.Info("stopping http service: " + name)
		sign <- 0
	}
	return nil
}

const UnlockService = "UNLOCK_SOCKET_SERVICES"
const ControlService = "CONTROL_SOCKET_SERVICES"

func (s *LegacyBootstrap) CreateConfig() {
	database := s.db

	err := database.AutoMigrate(&entity.SocketServerConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database: %v", err)
		return
	}
	var cnt int64
	database.Model(&entity.SocketServerConfig{}).Count(&cnt)
	if cnt == 0 {
		cert, key, err := secure.GenerateX509Cert()
		if err != nil {
			logger.Errorf("failed to generate x509 cert: %v", err)
			return
		}
		strCert := base64.StdEncoding.EncodeToString(cert)
		strKey := base64.StdEncoding.EncodeToString(key)
		socketServerConfig := entity.SocketServerConfig{
			ServiceName: UnlockService,
			Enable:      false,
			Host:        "0.0.0.0",
			Port:        2084,
			Cer:         strCert,
			Key:         strKey,
		}
		database.Create(&socketServerConfig)
		controlServerConfig := entity.SocketServerConfig{
			ServiceName: ControlService,
			Enable:      false,
			Host:        "0.0.0.0",
			Port:        2090,
			Cer:         strCert,
			Key:         strKey,
		}
		database.Create(&controlServerConfig)
	}

}
