package bootstrap

import (
	"encoding/base64"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/entity"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/secure"
	"fadacontrol/pkg/utils"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
	"sync"
	"time"
)

type DataInitBootstrap struct {
	_db         *gorm.DB
	adapter     *gormadapter.Adapter
	enforcer    *casbin.Enforcer
	_exitSignal *conf.ExitChanStruct
	startOnce   sync.Once
}

const HttpServiceApi = "HTTP_SERVICE_API"
const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func NewDataInitBootstrap(_exitSignal *conf.ExitChanStruct, adapter *gormadapter.Adapter, enforcer *casbin.Enforcer, _db *gorm.DB) *DataInitBootstrap {
	return &DataInitBootstrap{_exitSignal: _exitSignal, _db: _db, adapter: adapter, enforcer: enforcer}

}
func (d *DataInitBootstrap) Stop() error {
	return nil
}
func (d *DataInitBootstrap) Start() error {
	d.initUser()
	d.initHttpConfig()

	d.initRemoteConfig()
	d.initUdpConfig()
	d.initCasbinConfig()

	d.initSysConfig()

	return nil
}
func (d *DataInitBootstrap) initSysConfig() {
	d.startOnce.Do(func() {
		err := d._db.AutoMigrate(&entity.SysConfig{})
		if err != nil {
			logger.Errorf("failed to migrate database")
			return
		}
		var cnt int64
		err = d._db.Model(&entity.SysConfig{}).Count(&cnt).Error
		if err != nil {
			logger.Errorf("failed to count database")
			return
		}
		if cnt == 0 {
			sysConfig := entity.SysConfig{
				PowerSavingMode: true,
				Language:        "en",
				Region:          int(version.RegionGlobal),
			}
			d._db.Create(&sysConfig)
			goroutine.RecoverGO(func() {
				client, err := utils.NewClientBuilder().SetTimeout(5 * time.Second).Build()
				if err != nil {
					logger.Errorf("failed to create client")
					return
				}
				_, err = client.Get("https://www.google.com/")
				if err != nil {
					sysConfig.Region = int(version.RegionCN)
				} else {
					sysConfig.Region = int(version.RegionGlobal)
				}
				d._db.Save(&sysConfig)
			})

		}
	})

}
func (d *DataInitBootstrap) initHttpConfig() {

	err := d._db.AutoMigrate(&entity.HttpConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")
		return
	}
	var httpCount int64
	d._db.Model(&entity.HttpConfig{}).Count(&httpCount)
	if httpCount == 0 {
		cert, key, err := secure.GenerateX509Cert()
		if err != nil {
			logger.Errorf("failed to generate x509 cert: %v", err)
			return
		}
		strCert := base64.StdEncoding.EncodeToString(cert)
		strKey := base64.StdEncoding.EncodeToString(key)
		httpConfig := entity.HttpConfig{
			ServiceName: HttpServiceApi, //http service api
			Enable:      false,
			Host:        "0.0.0.0",
			Port:        2092,
			Cer:         "",
			Key:         "",
		}
		d._db.Create(&httpConfig)

		httpsConfig := entity.HttpConfig{
			ServiceName: HttpsServiceApi, //https service api
			Enable:      true,
			Host:        "0.0.0.0",
			Port:        2091,
			Cer:         strCert,
			Key:         strKey,
			EnableHttp3: false,
		}
		d._db.Create(&httpsConfig)

		httpAdminConfig := entity.HttpConfig{
			ServiceName: HttpServiceAdmin,
			Enable:      true,
			Host:        "127.0.0.1",
			Port:        2093,
			Cer:         "",
			Key:         "",
		}
		d._db.Create(&httpAdminConfig)
	}
}
func (d *DataInitBootstrap) initRemoteConfig() {
	err := d._db.AutoMigrate(&entity.RemoteConnectConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")

	}
	err = d._db.AutoMigrate(&entity.RemoteMsgServer{})
	if err != nil {
		logger.Errorf("failed to migrate database")
	}
	var count int64
	d._db.Model(&entity.RemoteConnectConfig{}).Count(&count)
	if count == 0 {
		key, err := secure.GenerateRandomBase58Key(35)
		if err != nil {
			logger.Errorf("failed to generate random base58 key")
		}

		remoteConfig := entity.RemoteConnectConfig{
			Enable:      false,
			SecurityKey: key,
		}
		d._db.Create(&remoteConfig)
	}

}
func (d *DataInitBootstrap) initUdpConfig() {
	err := d._db.AutoMigrate(&entity.DiscoverConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")
		return
	}
	var count int64
	d._db.Model(&entity.DiscoverConfig{}).Count(&count)
	if count == 0 {

		config := entity.DiscoverConfig{
			Enabled: true,
		}
		d._db.Create(&config)
	}

}
func (d *DataInitBootstrap) initCasbinConfig() {
	_, err := d.enforcer.AddPolicy("root", "*", "*")
	if err != nil {
		logger.Errorf("failed to add default policy")
		return
	}
	d.enforcer.AddPolicy("*", "api:/api/v1//unlock", "*")
	err = d.enforcer.SavePolicy()
	if err != nil {
		logger.Errorf("failed to save policy")
		return
	}

}
func (d *DataInitBootstrap) initUser() {
	err := d._db.AutoMigrate(&entity.User{})
	if err != nil {
		logger.Errorf("failed to migrate database")
		return
	}
	var count int64

	d._db.Model(&entity.User{}).Count(&count)
	if count == 0 {
		salt, _ := secure.GenerateSaltBase64(10)
		user := entity.User{
			Username: "root",
			Password: secure.HashPasswordByKDFBase64(conf.RootPassword, salt),
			Salt:     salt,
		}
		d._db.Create(&user)
	}
	if conf.ResetPassword {
		user := entity.User{}
		d._db.First(&user)
		salt, _ := secure.GenerateSaltBase64(10)
		user.Password = secure.HashPasswordByKDFBase64(conf.RootPassword, salt)
		user.Salt = salt
		d._db.Save(&user)
		logger.Info("root password reset")
		goroutine.RecoverGO(
			func() {
				d._exitSignal.ExitChan <- 0
			})
	}

}
