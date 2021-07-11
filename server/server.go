package server

import (
	"dhcp/models"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"net"
	"time"
)

var object *models.Object



func search(action, clientHW string, sign log.Fields) bool {
	var acls []models.ACL
	if err := object.Db.Where("client_hw_addr = ? and action = ?", clientHW, action).Find(&acls).Error; err != nil {
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

func DHCPD(d *DHCPDConfig, logLevel logger.LogLevel, connMaxLifetime time.Duration) {
	object = models.MustConnectDB(d.DBUser, d.DBHost, d.DBPass, d.DBName, d.DBPort, logLevel, d.DBPoolMaxIdleConns, d.DBPoolMaxOpenConns, connMaxLifetime)

	laddr := net.UDPAddr{
		IP:   net.ParseIP(d.Listen),
		Port: d.Port,
	}

	server, err := server4.NewServer(d.IFName, &laddr, handler)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
