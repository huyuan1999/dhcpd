package main

import (
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
)

var db *gorm.DB

func connectDB() {
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
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
	connectDB()
	err := db.AutoMigrate(&Leases{}, &Options{}, &ACL{}, &Binding{}, &Reserves{})
	if err != nil {
		log.Fatalln("auto migrate ", err.Error())
	}
}

func QueryOptions() Options {
	return Options{
		LeaseTime:    "1h",
		ServerIP:     "10.1.1.1",
		BootFileName: "pxelinux.0",
		GatewayIP:    "10.1.1.1",
		RangeStartIP: "10.1.1.100",
		RangeEndIP:   "10.1.1.110",
		NetMask:      "255.0.0.0",
		Router:       "10.1.1.2, 10.1.1.3",
		DNS:          "10.1.1.4, 10.1.1.5",
	}
}

func handler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	fields := log.Fields{
		"client_hw_addr": msg.ClientHWAddr,
		"transaction_id": msg.TransactionID,
		"hw_type":        msg.HWType,
		"message_type":   msg.MessageType(),
	}

	reply, err := dhcpv4.NewReplyFromRequest(msg)
	if err != nil {
		log.WithFields(fields).Errorf("NewReplyFromRequest %s", err.Error())
	}
	options := QueryOptions()
	switch msg.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		if err := Handler(conn, peer, reply, dhcpv4.MessageTypeOffer, options, fields); err != nil {
			log.WithFields(fields).Errorf("Generate offer reply message %s", err.Error())
		}
	case dhcpv4.MessageTypeRequest:
		if err := Handler(conn, peer, reply, dhcpv4.MessageTypeAck, options, fields); err != nil {
			log.WithFields(fields).Errorf("Generate ack reply message %s", err.Error())
		}
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
