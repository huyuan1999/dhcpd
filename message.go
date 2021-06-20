package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"log"
	"net"
	"time"
)

func checkValidNetmask(netmask net.IPMask) bool {
	netmaskInt := binary.BigEndian.Uint32(netmask)
	x := ^netmaskInt
	y := x + 1
	return (y & x) == 0
}

func getNetmask(ipMask string) (net.IPMask, error) {
	netmaskIP := net.ParseIP(ipMask)
	if netmaskIP.IsUnspecified() {
		return nil, errors.New("invalid subnet mask")
	}

	netmaskIP = netmaskIP.To4()
	if netmaskIP == nil {
		return nil, errors.New("error converting subnet mask to IPv4 format")
	}

	netmask := net.IPv4Mask(netmaskIP[0], netmaskIP[1], netmaskIP[2], netmaskIP[3])
	if !checkValidNetmask(netmask) {
		return nil, errors.New("illegal subnet mask")
	}
	return netmask, nil
}

func Offer(msg *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, error) {
	LeaseTime, err := time.ParseDuration("1h")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("lease generation time error %s", err))
	}
	msg.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
	msg.YourIPAddr = net.ParseIP("10.1.1.100")
	msg.ServerIPAddr = net.ParseIP("10.1.1.1")
	msg.GatewayIPAddr = net.ParseIP("10.1.1.1")
	msg.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP("10.1.1.1")))
	msg.UpdateOption(dhcpv4.OptIPAddressLeaseTime(LeaseTime))
	msg.BootFileName = "pxelinux.0"
	subnetMask, err := getNetmask("255.0.0.0")
	if err != nil {
		log.Fatalln(err)
	}
	msg.UpdateOption(dhcpv4.OptSubnetMask(subnetMask))
	return msg, nil
}

func ACK(msg *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, error) {
	LeaseTime, err := time.ParseDuration("1h")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("lease generation time error %s", err))
	}
	msg.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	msg.YourIPAddr = net.ParseIP("10.1.1.100")
	msg.ServerIPAddr = net.ParseIP("10.1.1.1")
	msg.GatewayIPAddr = net.ParseIP("10.1.1.1")
	msg.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP("10.1.1.1")))
	msg.UpdateOption(dhcpv4.OptIPAddressLeaseTime(LeaseTime))
	msg.BootFileName = "pxelinux.0"
	subnetMask, err := getNetmask("255.0.0.0")
	if err != nil {
		log.Fatalln(err)
	}
	msg.UpdateOption(dhcpv4.OptSubnetMask(subnetMask))
	return msg, nil
}


