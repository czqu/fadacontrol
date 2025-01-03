package bootstrap

import (
	"context"
	"encoding/base64"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	_log "fadacontrol/internal/base/log"
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
	"os"
	"sync"
	"time"
)

type DataInitBootstrap struct {
	_db      *gorm.DB
	adapter  *gormadapter.Adapter
	enforcer *casbin.Enforcer

	startOnce sync.Once
	ctx       context.Context
}

const HttpsServiceApi = "HTTPS_SERVICE_API"
const HttpServiceAdmin = "HTTP_SERVICE_ADMIN"

func NewDataInitBootstrap(ctx context.Context, adapter *gormadapter.Adapter, enforcer *casbin.Enforcer, _db *gorm.DB) *DataInitBootstrap {
	return &DataInitBootstrap{ctx: ctx, _db: _db, adapter: adapter, enforcer: enforcer}

}
func (d *DataInitBootstrap) Stop() error {
	return nil
}
func (d *DataInitBootstrap) Start() error {
	d.initSysConfig()
	d.initLogReport()
	d.initUser()
	d.initHttpConfig()
	d.initRemoteConfig()
	d.initUdpConfig()
	d.initCasbinConfig()
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
		logger.InitLogReporter(_log.NewDefaultSentryOptions())
		logger.Fatal("failed to migrate database")
		return
	}
	opt := utils.GetLogReporterOPtions(region)
	defer func() {
		_conf := utils.GetValueFromContext(d.ctx, constants.ConfKey, conf.NewDefaultConf())
		_conf.LogReporterOpt = opt
	}()
	var cnt int64
	err = d._db.Model(&entity.LogReportSentry{}).Count(&cnt).Error
	if err != nil {
		logger.Errorf("failed to count database")
		return
	}
	if cnt == 0 {
		sentryConfig := entity.LogReportSentry{
			Enable:      opt.Enable,
			UserId:      uuid.New().String(),
			ReportLevel: opt.Level,
		}
		d._db.Create(&sentryConfig)
	}
	sentryConfig := entity.LogReportSentry{}
	err = d._db.First(&sentryConfig).Error
	if err != nil {
		logger.Errorf("failed to get config %v", err)
		logger.InitLogReporter(opt)
		return
	}
	opt.UserId = sentryConfig.UserId
	opt.Enable = sentryConfig.Enable
	logger.InitLogReporter(opt)

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
		logger.Error("failed to add default policy", err)
	}

	_, err = d.enforcer.AddPolicy("*", "api:/api/v1/unlock", "*")
	if err != nil {
		logger.Error("failed to add default policy", err)
	}

	d._db.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'casbin_rule'")
	err = d.enforcer.SavePolicy()
	if err != nil {
		logger.Error("failed to save policy", err)
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
				cancelFunc := d.ctx.Value(constants.CancelFuncKey).(context.CancelFunc)
				if cancelFunc != nil {
					cancelFunc()
				} else {
					os.Exit(0)
				}
			})
	}

}
