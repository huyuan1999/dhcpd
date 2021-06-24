package main

import (
	"encoding/binary"
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

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
	ok, err := checkLeases(addr, clientHW)
	if err != nil {
		log.Errorf("Error checkIfTaken call checkLeases %s", err.Error())
	}
	return ok
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
			return nil, errors.New("no new ip addresses available")
		}
		taken = checkIfTaken(ip, clientHW)
	}
	return ip, nil
}

// 如果 addr 存在且 clientHW 相同则更新租约到期时间，并返回 false
// 如果 addr 存在且 clientHW 不同则返回 true ，表示此 ip 地址已经被分配
// 如果 addr 不存在则表示此 ip 地址尚未被分配，将租约信息写入到数据库，并返回 false
func checkLeases(addr string, clientHW string) (ok bool, err error) {
	var lease Leases

	options := QueryOptions()

	leaseTime, err := time.ParseDuration(options.LeaseTime)
	if err != nil {
		return true, errors.New(fmt.Sprintf("lease generation time %s", err.Error()))
	}

	// addr 存在且 clientHW 相同
	if err := db.Where("assigned_addr = ? and client_hw_addr = ?", addr, clientHW).First(&lease).Error; err == nil {
		lease.Expires = time.Now().Add(leaseTime)
		if err := db.Save(&lease).Error; err != nil {
			return true, errors.New(fmt.Sprintf("update lease expires %s", err.Error()))
		}
		return false, nil
	}

	// addr 存在且 clientHW 不同（表示此地址已经被分配给别的主机）
	if err := db.Where("assigned_addr = ?", addr).First(&lease).Error; err == nil {
		return true, nil
	}

	lease.Expires = time.Now().Add(leaseTime)
	lease.AssignedAddr = addr
	lease.ClientHWAddr = clientHW
	if err := db.Create(&lease).Error; err != nil {
		return true, errors.New(fmt.Sprintf("create lease info %s", err.Error()))
	}
	return false, nil
}

// 分配一个IP地址给客户端
func createIP(rangeStart string, rangeEnd string, clientHW string) (net.IP, error) {
	var bind Binding
	var lease Leases

	// 检查这个客户端是否有绑定的IP地址
	if err := db.Where("client_hw_addr = ?", clientHW).First(&bind).Error; err == nil {
		ok, err := checkLeases(bind.BindAddr, clientHW)
		if err != nil {
			return nil, errors.Wrap(err, "checkLeases")
		}
		// 如果 checkLeases 返回 true, 且 err 为 nil 则表示绑定的 IP 地址被分配了给其他机器
		if ok {
			return nil, errors.New("the bound IP address is assigned to another machine")
		}
		return net.ParseIP(bind.BindAddr), nil
	}

	// 检查这个客户端是否已经分配了IP地址(如果已经分配则按照续约请求处理)
	if err := db.Where("client_hw_addr = ?", clientHW).First(&lease).Error; err == nil {
		options := QueryOptions()
		leaseTime, err := time.ParseDuration(options.LeaseTime)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("lease generation time %s", err.Error()))
		}

		lease.Expires = time.Now().Add(leaseTime)
		if err := db.Save(&lease).Error; err != nil {
			return nil, errors.New(fmt.Sprintf("update lease info %s", err.Error()))
		}
		return net.ParseIP(lease.AssignedAddr), nil
	}
	return assignedIP(rangeStart, rangeEnd, clientHW)
}

func Handler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4, messageType dhcpv4.MessageType, options *Options, fields log.Fields) {
	// 设置租约时间
	leaseTime, err := time.ParseDuration(options.LeaseTime)
	if err != nil {
		log.WithFields(fields).Errorf("Error lease generation time error %s", err.Error())
		return
	}

	// 获取将要分配给客户端的地址
	assignedIP, err := createIP(options.RangeStartIP, options.RangeEndIP, msg.ClientHWAddr.String())
	if err != nil {
		log.WithFields(fields).Errorf("Error create IP assigned to client %s", err.Error())
		return
	}

	// 解析子网掩码
	subnetMask, err := getNetmask(options.NetMask)
	if err != nil {
		log.WithFields(fields).Errorf("Error parsing subnet mask %s", err.Error())
		return
	}

	router := parse(options.Router)
	dns := parse(options.DNS)

	// 构建 dhcp 响应包
	msg.UpdateOption(dhcpv4.OptMessageType(messageType))
	msg.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP(options.ServerIP)))
	msg.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseTime))
	msg.UpdateOption(dhcpv4.OptSubnetMask(subnetMask))
	msg.UpdateOption(dhcpv4.OptRouter(router...))
	msg.UpdateOption(dhcpv4.OptDNS(dns...))
	msg.BootFileName = options.BootFileName
	msg.YourIPAddr = assignedIP
	msg.ServerIPAddr = net.ParseIP(options.ServerIP)
	msg.GatewayIPAddr = net.ParseIP(options.GatewayIP)

	if _, err := conn.WriteTo(msg.ToBytes(), peer); err != nil {
		log.WithFields(fields).Errorf("Error Write DHCP reply message %s", err.Error())
	}
}

func ReleaseHandler(msg *dhcpv4.DHCPv4, fields log.Fields) {
	withReleaseAddress(msg, fields, "ReleaseHandler")
}

func DeclineHandler(msg *dhcpv4.DHCPv4, fields log.Fields) {
	withReleaseAddress(msg, fields, "DeclineHandler")
}

func withReleaseAddress(msg *dhcpv4.DHCPv4, fields log.Fields, handlerName string) {
	var leases Leases
	clientHWAddr := msg.ClientHWAddr.String()
	if err := db.Unscoped().Where("client_hw_addr = ?", clientHWAddr).Delete(&leases).Error; err != nil {
		log.WithFields(fields).Warningf("%s release address %s", handlerName, err.Error())
	}
}
