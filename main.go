package main

import (
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"log"
	"net"
)

func handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	log.Println(m.Summary())
	tmp, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		log.Fatalln("error in constructing offer response message", err)
	}

	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		msg, err := Offer(tmp)
		if err != nil {
			log.Println(err)
			return
		}
		n, err := conn.WriteTo(msg.ToBytes(), peer)
		if err != nil {
			log.Fatalln("error writing offer response message", err)
		}
		log.Println("write offer package successfully: ", n)
	case dhcpv4.MessageTypeRequest:
		msg, err:= ACK(tmp)
		if err != nil {
			log.Println(err)
			return
		}
		n, err := conn.WriteTo(msg.ToBytes(), peer)
		if err != nil {
			log.Fatalln("error writing offer response message", err)
		}
		log.Println("write ack package successfully: ", n)
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
