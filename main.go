package main

import (
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"sync"
)

var (
	db               *gorm.DB
	OptionsCache     *Options
	OptionsCacheLock sync.Mutex
)

func MustConnectDB() {
	var err error
	dsn := "root:123456@tcp(192.168.3.50:3306)/dhcpd?charset=utf8&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalln("connect to database ", err.Error())
	}
}

func init() {
	MustConnectDB()
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
	err := db.AutoMigrate(&Leases{}, &Options{}, &ACL{}, &Binding{}, &Reserves{})
	if err != nil {
		log.Fatalf("db auto migrate %s", err.Error())
	}
}

// 从数据库查询配置, 如果查询出现错误则读取上次查询的结果
// 如果上次查询的结果为 nil 则退出程序并输出相关日志
func QueryOptions() *Options {
	var options Options

	if err := db.First(&options).Error; err != nil {
		log.Errorf("QueryOptions %s", err.Error())
		if OptionsCache != nil {
			return OptionsCache
		}
		log.Fatalf("QueryOptions %s and OptionsCache is nil", err.Error())
	}
	OptionsCacheLock.Lock()
	OptionsCache = &options
	OptionsCacheLock.Unlock()
	return &options
}

func acl(clientHW string) bool {
	return false
}

func handler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	fields := log.Fields{
		"client_hw_addr": msg.ClientHWAddr,
		"transaction_id": msg.TransactionID,
		"hw_type":        msg.HWType,
		"message_type":   msg.MessageType(),
	}

	if acl(msg.ClientHWAddr.String()) {
		log.WithFields(fields).Infoln("acl rule mismatch")
		return
	}

	reply, err := dhcpv4.NewReplyFromRequest(msg)
	if err != nil {
		log.WithFields(fields).Errorf("New reply from request %s", err.Error())
	}

	options := QueryOptions()

	switch msg.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		Handler(conn, peer, reply, dhcpv4.MessageTypeOffer, options, fields)
	case dhcpv4.MessageTypeRequest:
		Handler(conn, peer, reply, dhcpv4.MessageTypeAck, options, fields)
	case dhcpv4.MessageTypeDecline:
		DeclineHandler(reply, fields)
	case dhcpv4.MessageTypeRelease:
		ReleaseHandler(reply, fields)
	default:
		log.WithFields(fields).Infoln("An unknown request was received")
	}
}

func main() {
	laddr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 67,
	}
	server, err := server4.NewServer("em2", &laddr, handler)
	if err != nil {
		log.Fatal(err)
	}

	server.Serve()
}
