package models

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Object struct {
	Db   *gorm.DB
	Sqlx *sql.DB
}

func ConnectDB(user, host, password, name string, port int, logLevel logger.LogLevel, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) (*Object, error) {
	object := &Object{}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, name)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{Logger: logger.Default.LogMode(logLevel),})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	}
	object.Db = db
	object.Sqlx = sqlDB
	return object, nil
}

func MustConnectDB(user, host, password, name string, port int, logLevel logger.LogLevel, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) *Object {
	object, err := ConnectDB(user, host, password, name, port, logLevel, maxIdleConns, maxOpenConns, connMaxLifetime)
	if err != nil {
		panic(err)
	}
	return object
}
