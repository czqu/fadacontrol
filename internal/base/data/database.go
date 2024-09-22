package data

import (
	"errors"
	"fadacontrol/internal/base/conf"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Data struct {
	DB *gorm.DB
}

func NewData(db *gorm.DB) *Data {
	return &Data{
		DB: db,
	}
}
func NewDB(dataConf *conf.DatabaseConf) (*gorm.DB, error) {
	if dataConf.Driver != "sqlite" {
		return nil, errors.New("Unsupported driver")
	}
	config := &gorm.Config{}
	if dataConf.Debug {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}
	engine, err := gorm.Open(sqlite.Open(dataConf.Connection), config)
	if err != nil {
		return nil, err
	}
	sqlDB, dbError := engine.DB()
	if dbError != nil {
		return nil, fmt.Errorf("failed to create sqlDB")
	}

	sqlDB.SetMaxIdleConns(dataConf.MaxIdleConnection)
	sqlDB.SetMaxOpenConns(dataConf.MaxOpenConnection)

	return engine, nil
}
