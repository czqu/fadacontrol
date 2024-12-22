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
	"github.com/google/uuid"
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

const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func NewDataInitBootstrap(_exitSignal *conf.ExitChanStruct, adapter *gormadapter.Adapter, enforcer *casbin.Enforcer, _db *gorm.DB) *DataInitBootstrap {
	return &DataInitBootstrap{_exitSignal: _exitSignal, _db: _db, adapter: adapter, enforcer: enforcer}

}
func (d *DataInitBootstrap) Stop() error {
	return nil
}
func (d *DataInitBootstrap) Start() error {
	d.initLogReport()
	d.initUser()
	d.initHttpConfig()
	d.initRemoteConfig()
	d.initUdpConfig()
	d.initCasbinConfig()
	d.initSysConfig()
	return nil
}
func (d *DataInitBootstrap) initLogReport() {

	sysconfig := entity.SysConfig{}
	region := version.RegionGlobal
	if err := d._db.First(&sysconfig).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		region = version.GetRegionFromCode(sysconfig.Region)
	}
	err := d._db.AutoMigrate(&entity.LogReportSentry{})
	if err != nil {
		logger.InitLogReporter(logger.NewDefaultSentryOptions(), "fatal")
		logger.Fatal("failed to migrate database")
		return
	}
	enableReport, _ := utils.GetRemoteConfig("log_report_enable", region, true)
	reportLevel, _ := utils.GetRemoteConfig("log_report_min_level", region, "info")
	profilesSampleRate, _ := utils.GetRemoteConfig("log_report_sentry_profiles_sample_rate", region, 0.2)
	tracesSampleRate, _ := utils.GetRemoteConfig("log_report_sentry_traces_sample_rate", region, 0.2)
	var cnt int64
	err = d._db.Model(&entity.LogReportSentry{}).Count(&cnt).Error
	if err != nil {
		logger.Errorf("failed to count database")
		return
	}
	if cnt == 0 {
		sentryConfig := entity.LogReportSentry{
			Enable:      enableReport.(bool),
			UserId:      uuid.New().String(),
			ReportLevel: reportLevel.(string),
		}
		d._db.Create(&sentryConfig)
	}
	sentryConfig := entity.LogReportSentry{}
	err = d._db.First(&sentryConfig).Error
	if err != nil {
		logger.Errorf("failed to get config %v", err)
		logger.InitLogReporter(logger.NewDefaultSentryOptions(), reportLevel.(string))
		return
	}
	options := &logger.SentryOptions{
		UserId:             sentryConfig.UserId,
		TracesSampleRate:   tracesSampleRate.(float64),
		ProfilesSampleRate: profilesSampleRate.(float64),
	}
	if !enableReport.(bool) {
		logger.InitLogReporter(options, reportLevel.(string))
	} else {
		logger.InitLogReporter(options, reportLevel.(string))
	}

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
		logger.Warnf("Failed to migrate database: %v. Attempting to drop and recreate the table.", err)
		if dropErr := d._db.Migrator().DropTable(&entity.HttpConfig{}); dropErr != nil {
			logger.Errorf("Failed to drop table: %v", dropErr)
			return
		}
		if recreateErr := d._db.AutoMigrate(&entity.HttpConfig{}); recreateErr != nil {
			logger.Errorf("Failed to recreate table: %v", recreateErr)
			return
		}

		logger.Infof("Table recreated successfully.")
	}
	var config1 entity.HttpConfig
	adminServiceErr := d._db.Where(&entity.HttpConfig{ServiceName: HttpServiceAdmin}).First(&config1).Error
	var config2 entity.HttpConfig
	commonServiceErr := d._db.Where(&entity.HttpConfig{ServiceName: HttpsServiceApi}).First(&config2).Error

	var httpCount int64
	d._db.Model(&entity.HttpConfig{}).Count(&httpCount)
	if httpCount != 2 || adminServiceErr != nil || commonServiceErr != nil {

		if dropErr := d._db.Migrator().DropTable(&entity.HttpConfig{}); dropErr != nil {
			logger.Errorf("Failed to drop table: %v", dropErr)
			return
		}
		if recreateErr := d._db.AutoMigrate(&entity.HttpConfig{}); recreateErr != nil {
			logger.Errorf("Failed to recreate table: %v", recreateErr)
			return
		}

		logger.Infof("Table recreated successfully.")

		cert, key, err := secure.GenerateX509Cert()
		if err != nil {
			logger.Errorf("failed to generate x509 cert: %v", err)
			return
		}
		strCert := base64.StdEncoding.EncodeToString(cert)
		strKey := base64.StdEncoding.EncodeToString(key)

		httpsConfig := entity.HttpConfig{
			ServiceName: HttpsServiceApi, //https service api
			Enable:      true,
			Host:        "0.0.0.0",
			Port:        2091,
			Cer:         strCert,
			Key:         strKey,
			EnableHttp3: false,
		}

		d._db.Save(&httpsConfig)

		httpAdminConfig := entity.HttpConfig{
			ServiceName: HttpServiceAdmin,
			Enable:      true,
			Host:        "127.0.0.1",
			Port:        2093,
			Cer:         "",
			Key:         "",
		}

		d._db.Save(&httpAdminConfig)
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
