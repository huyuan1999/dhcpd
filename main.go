package main

import (
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
	db *gorm.DB
)

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

func MustConnectDB() {
	var err error
	dsn := "root:123456@tcp(192.168.3.50:3306)/dhcpd?charset=utf8&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Info),})
	if err != nil {
		log.Fatalln("connect to database ", err.Error())
		return
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return
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

func search(s []ACL, clientHW string) bool {
	for _, item := range s {
		if item.ClientHWAddr == clientHW {
			return true
		}
	}
	return false
}

func acl(clientHW string, sign log.Fields) bool {
	var acls []ACL
	options := QueryOptions()
	// 是否打开 ACL 控制
	if !options.ACL {
		return false
	}

	switch options.ACLAction {
	case "allow":
		// 如果查询数据库发送错误则直接返回 true (禁止为此客户端分配IP)
		if err := db.Where("client_hw_addr = ? and action = ?", clientHW, "allow").Find(&acls).Error; err != nil {
			log.WithFields(sign).Errorf("Error query acl %s", err.Error())
			return true
		}
		if search(acls, clientHW) {
			return false
		}
		return true
	case "deny":
		// 如果查询数据库发送错误则直接返回 true (禁止为此客户端分配IP)
		if err := db.Where("client_hw_addr = ? and action = ?", clientHW, "deny").Find(&acls).Error; err != nil {
			log.WithFields(sign).Errorf("Error query acl %s", err.Error())
			return true
		}
		if search(acls, clientHW) {
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

	// 返回 true 则表示禁止为此客户端分配IP地址
	if acl(msg.ClientHWAddr.String(), sign) {
		log.WithFields(sign).Infoln("acl rule mismatch")
		return
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

func main() {
	go DeleteExpiredLease()

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
