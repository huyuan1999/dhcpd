package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"strings"
	"time"
)

// 配置信息(从数据库中读取)
type Options struct {
	LeaseTime    string
	ServerIP     string
	BootFileName string
	GatewayIP    string
	RangeStartIP string
	RangeEndIP   string
	NetMask      string
	Router       string
	DNS          string
}

// 检查IP是否已被分配, 返回true表示已分配
func checkIfTaken(ip net.IP) bool {
	return false
}

// 从可分配的IP地址返回随机获取一个可用的IP地址
func createIP(rangeStart net.IP, rangeEnd net.IP) (net.IP, error) {
	ip := make([]byte, 4)
	rangeStartInt := binary.BigEndian.Uint32(rangeStart.To4())
	rangeEndInt := binary.BigEndian.Uint32(rangeEnd.To4())
	binary.BigEndian.PutUint32(ip, random(rangeStartInt, rangeEndInt))
	taken := checkIfTaken(ip)
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt++
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt > rangeEndInt {
			break
		}
		taken = checkIfTaken(ip)
	}
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt--
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt < rangeStartInt {
			return nil, errors.New("no new IP addresses available")
		}
		taken = checkIfTaken(ip)
	}
	return ip, nil
}

func random(min uint32, max uint32) uint32 {
	return uint32(rand.Intn(int(max-min))) + min
}

// 将以逗号(,)分隔的IP列表转换为 []net.IP 对象
func parse(addrs string) (ips []net.IP) {
	s := strings.Split(addrs, ",")
	for _, addr := range s {
		ips = append(ips, net.ParseIP(strings.Trim(addr, " ")))
	}
	return ips
}

func Handler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4, messageType dhcpv4.MessageType, options Options, fields log.Fields) error {
	// 设置租约时间
	LeaseTime, err := time.ParseDuration(options.LeaseTime)
	if err != nil {
		return errors.New(fmt.Sprintf("lease generation time error %s", err))
	}

	// 获取将要分配给客户端的地址
	assignedIP, err := createIP(net.ParseIP(options.RangeStartIP), net.ParseIP(options.RangeEndIP))
	if err != nil {
		return err
	}

	// 解析子网掩码
	subnetMask, err := getNetmask(options.NetMask)
	if err != nil {
		return err
	}

	router := parse(options.Router)
	dns := parse(options.DNS)

	// 构建 dhcp 响应包
	msg.UpdateOption(dhcpv4.OptMessageType(messageType))
	msg.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP(options.ServerIP)))
	msg.UpdateOption(dhcpv4.OptIPAddressLeaseTime(LeaseTime))
	msg.UpdateOption(dhcpv4.OptSubnetMask(subnetMask))
	msg.UpdateOption(dhcpv4.OptRouter(router...))
	msg.UpdateOption(dhcpv4.OptDNS(dns...))
	msg.BootFileName = options.BootFileName
	msg.YourIPAddr = assignedIP
	msg.ServerIPAddr = net.ParseIP(options.ServerIP)
	msg.GatewayIPAddr = net.ParseIP(options.GatewayIP)

	if _, err := conn.WriteTo(msg.ToBytes(), peer); err != nil {
		log.WithFields(fields).Errorf("Write DHCP reply message %s", err.Error())
	}

	return nil
}
