package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net"
	"strings"
	"testing"
)

type Options2 struct {
	LeaseTime    string
	ServerIP     string
	BootFileName string
	GatewayIP    string
	RangeStartIP string
	RangeEndIP   string
	NetMask      string
	Router       []byte
	DNS          []byte
}

func connect() *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root:123456@tcp(192.168.3.50:3306)/dhcpd?charset=utf8&parseTime=True&loc=Local", // DSN data source name
		DefaultStringSize:         256,                                                                              // string 类型字段的默认长度
		DisableDatetimePrecision:  true,                                                                             // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,                                                                             // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,                                                                             // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,                                                                            // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})
	if err != nil {
		log.Fatalln("连接数据库错误")
	}
	return db
}

func TestHandler(t *testing.T) {
	//db := connect()
	//t.Log(db.AutoMigrate(&Options2{}))

	//x := []net.IP{net.ParseIP("10.1.1.1"), net.ParseIP("10.1.1.2"), net.ParseIP("10.1.1.3"),}
	//y, _ := json.Marshal(x)
	//t.Log(string(y))
	addrs := "10.1.1.4, 10.1.1.5, 10.1.1.6"
	s := strings.Split(addrs, ",")
	for _, i := range(s) {
		t.Log(net.ParseIP(strings.Trim(i, " ")))
	}
	t.Log(net.ParseIP("1.1.1.1"))
}
