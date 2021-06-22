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

// 如果 addr 存在且 clientHW 相同则更新租约到期时间，并返回 false
// 如果 addr 存在且 clientHW 不同则返回 true，表示此 ip 地址已经被分配
// 如果 addr 不存在则表示此 ip 地址尚未被分配，将租约信息写入到数据库，并返回 false
func checkLeases(addr string, clientHW string) bool {
	var lease Leases
	var options Options

	if err := db.First(&options).Error; err != nil {
		log.Errorln("first lease info ", err.Error())
		return true
	}

	LeaseTime, err := time.ParseDuration(options.LeaseTime)
	if err != nil {
		log.Errorln("lease generation time error ", err.Error())
		return true
	}

	if err := db.Where("assigned_addr = ? and client_hw_addr = ?", addr, clientHW).First(&lease).Error; err == nil {
		lease.Expires = time.Now().Add(LeaseTime)
		if err := db.Save(&lease).Error; err != nil {
			log.Errorln("update lease expires ", err.Error())
			return true
		}
		return false
	}

	// addr 存在且 clientHW 不同
	if err := db.Where("assigned_addr = ?", addr).First(&lease).Error; err == nil {
		return true
	}

	lease.Expires = time.Now().Add(LeaseTime)
	lease.AssignedAddr = addr
	lease.ClientHWAddr = clientHW
	if err := db.Create(&lease).Error; err != nil {
		log.Errorln("create lease info ", err.Error())
		return true
	}

	return false
}

// 检查IP是否已被分配, 返回true表示已分配
func checkIfTaken(ip net.IP, clientHW string) bool {
	var bind Binding
	var reserve Reserves
	addr := ip.String()
	if err := db.Where("bind_addr = ?", addr).First(&bind).Error; err == nil {
		return true
	}

	if err := db.Where("address = ?", addr).First(&reserve).Error; err == nil {
		return true
	}
	return checkLeases(addr, clientHW)
}

// 从可分配的IP地址返回随机获取一个可用的IP地址
func assignedIP(rangeStart string, rangeEnd string, clientHW string) (net.IP, error) {
	ip := make([]byte, 4)
	start := net.ParseIP(rangeStart)
	end := net.ParseIP(rangeEnd)
	rangeStartInt := binary.BigEndian.Uint32(start.To4())
	rangeEndInt := binary.BigEndian.Uint32(end.To4())
	binary.BigEndian.PutUint32(ip, random(rangeStartInt, rangeEndInt))
	taken := checkIfTaken(ip, clientHW)
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt++
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt > rangeEndInt {
			break
		}
		taken = checkIfTaken(ip, clientHW)
	}
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt--
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt < rangeStartInt {
			return nil, errors.New("no new IP addresses available")
		}
		taken = checkIfTaken(ip, clientHW)
	}
	return ip, nil
}

// 分配一个IP地址给客户端
func createIP(rangeStart string, rangeEnd string, clientHW string) (net.IP, error) {
	var bind Binding
	var lease Leases
	var options Options
	// 检查这个客户端是否有绑定的IP地址
	if err := db.Where("client_hw_addr = ?", clientHW).First(&bind).Error; err == nil {
		if checkLeases(bind.BindAddr, clientHW) {
			// 如果 checkLeases 返回 true 则表示绑定的 IP 地址被分配了给其他机器
			return nil, errors.New("the bound IP address is assigned to another machine")
		}
		return net.ParseIP(bind.BindAddr), nil
	}
	// 检查这个客户端是否已经分配了IP地址(如果已经分配则按照续约请求处理)
	if err := db.Where("client_hw_addr = ?", clientHW).First(&lease).Error; err == nil {
		if err := db.First(&options).Error; err != nil {
			return nil, err
		}

		LeaseTime, err := time.ParseDuration(options.LeaseTime)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("lease generation time error ", err.Error()))
		}

		lease.Expires = time.Now().Add(LeaseTime)
		if err := db.Save(&lease).Error; err != nil {
			return nil, err
		}
		return net.ParseIP(lease.AssignedAddr), nil
	}

	return assignedIP(rangeStart, rangeEnd, clientHW)
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
	assignedIP, err := createIP(options.RangeStartIP, options.RangeEndIP, msg.ClientHWAddr.String())
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
