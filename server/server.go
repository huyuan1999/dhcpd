package server

import (
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"time"
)

var (
	db, Db *gorm.DB
)

// 每分钟检查一次租约表, 删除过期的租约信息
func DeleteExpiredLease() {
	c := cron.New()
	_, err := c.AddFunc("* * * * *", func() {
		options := QueryOptions()
		leaseTime, err := time.ParseDuration(options.LeaseTime)
		if err != nil {
			log.Fatalf("Error delete expired lease lease generation time %s", err.Error())
			return
		}
		db.Unscoped().Where("unix_timestamp(expires) < ?", time.Now().Add(leaseTime).Unix()).Delete(&Leases{})
	})
	if err != nil {
		log.Fatalf("Error init delete expired lease cron job %s", err.Error())
	}
	c.Start()
}

func MustConnectDB(user, host, password, name string, port int, logLevel logger.LogLevel, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, name)
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{Logger: logger.Default.LogMode(logLevel),})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
		return
	}
}

func search(action, clientHW string, sign log.Fields) bool {
	var acls []ACL
	if err := db.Where("client_hw_addr = ? and action = ?", clientHW, action).Find(&acls).Error; err != nil {
		log.WithFields(sign).Errorf("Error query acl %s", err.Error())
		return true
	}
	for _, item := range acls {
		if item.ClientHWAddr == clientHW {
			return true
		}
	}
	return false
}

func acl(clientHW string, sign log.Fields) bool {
	options := QueryOptions()
	// 是否打开 ACL 控制
	if !options.ACL {
		return false
	}

	switch options.ACLAction {
	case "allow":
		if search("allow", clientHW, sign) {
			return false
		}
		return true
	case "deny":
		log.WithFields(sign).Infoln("acl rule mismatch deny")
		if search("deny", clientHW, sign) {
			return true
		}
		return false
	default:
		log.WithFields(sign).Errorf("Error unknown acl action")
		return true
	}
}

func handler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	sign := log.Fields{
		"client_hw_addr": msg.ClientHWAddr,
		"transaction_id": msg.TransactionID,
		"hw_type":        msg.HWType,
		"message_type":   msg.MessageType(),
	}

	if msg.MessageType() == dhcpv4.MessageTypeDiscover || msg.MessageType() == dhcpv4.MessageTypeRequest {
		// 返回 true 则表示禁止为此客户端分配IP地址
		if acl(msg.ClientHWAddr.String(), sign) {
			return
		}
	}

	reply, err := dhcpv4.NewReplyFromRequest(msg)
	if err != nil {
		log.WithFields(sign).Errorf("New reply from request %s", err.Error())
	}

	switch msg.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		NewHandler(conn, peer, reply, dhcpv4.MessageTypeOffer, sign).OfferHandler()
	case dhcpv4.MessageTypeRequest:
		NewHandler(conn, peer, reply, dhcpv4.MessageTypeAck, sign).AckHandler()
	case dhcpv4.MessageTypeDecline:
		NewHandler(conn, peer, reply, dhcpv4.MessageTypeDecline, sign).DeclineHandler()
	case dhcpv4.MessageTypeRelease:
		NewHandler(conn, peer, reply, dhcpv4.MessageTypeRelease, sign).ReleaseHandler()
	default:
		log.WithFields(sign).Infoln("An unknown request was received")
	}
}

func initDB(user, host, password, name string, port int, logLevel logger.LogLevel, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) {
	MustConnectDB(user, host, password, name, port, logLevel, maxIdleConns, maxOpenConns, connMaxLifetime)
	err := db.AutoMigrate(&Leases{}, &Options{}, &ACL{}, &Binding{}, &Reserves{})
	if err != nil {
		log.Fatalf("db auto migrate %s", err.Error())
	}
	// 为 restful 暴露出去的全局变量
	Db = db
}

type DHCPDConfig struct {
	Listen                string
	Port                  int
	IFName                string
	Debug                 bool
	DBUser                string
	DBHost                string
	DBPass                string
	DBPort                int
	DBName                string
	DBPoolMaxIdleConns    int
	DBPoolMaxOpenConns    int
	DBPoolConnMaxLifetime int
}

func DHCPD(conf *DHCPDConfig, control chan struct{}) {
	log.SetReportCaller(true)
	log.SetLevel(log.WarnLevel)
	dbLogLevel := logger.Error
	if conf.Debug {
		dbLogLevel = logger.Info
		log.SetLevel(log.DebugLevel)
	}

	t := time.Second * time.Duration(conf.DBPoolConnMaxLifetime)

	initDB(conf.DBUser, conf.DBHost, conf.DBPass, conf.DBName, conf.DBPort, dbLogLevel, conf.DBPoolMaxIdleConns, conf.DBPoolMaxOpenConns, t)
	laddr := net.UDPAddr{
		IP:   net.ParseIP(conf.Listen),
		Port: conf.Port,
	}

	server, err := server4.NewServer(conf.IFName, &laddr, handler)
	if err != nil {
		panic(err)
	}

	go DeleteExpiredLease()

	// 通知其他 goroutine dhcpd 完成前置准备工作
	control <- struct{}{}

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
