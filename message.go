package main

import (
	"encoding/binary"
	"fmt"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var (
	OptionsCache     *Options
	OptionsCacheLock sync.Mutex
)

type Handler struct {
	conn        net.PacketConn
	peer        net.Addr
	msg         *dhcpv4.DHCPv4
	messageType dhcpv4.MessageType
	sign        log.Fields
	options     *Options
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

func NewHandler(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4, msgType dhcpv4.MessageType, sign log.Fields) *Handler {
	options := QueryOptions()
	return &Handler{
		conn:        conn,
		peer:        peer,
		msg:         msg,
		messageType: msgType,
		sign:        sign,
		options:     options,
	}
}

func (h *Handler) OfferHandler() {
	h.withReplyHandler()
}

func (h *Handler) AckHandler() {
	h.withReplyHandler()
}

func (h *Handler) withReplyHandler() {
	// 设置租约时间
	leaseTime, err := time.ParseDuration(h.options.LeaseTime)
	if err != nil {
		log.WithFields(h.sign).Errorf("Error lease generation time %s", err.Error())
		return
	}

	// 获取将要分配给客户端的地址
	assignedIP, err := h.createIP(h.options.RangeStartIP, h.options.RangeEndIP)
	if err != nil {
		log.WithFields(h.sign).Errorf("Error create IP assigned to client %s", err.Error())
		return
	}

	// 解析子网掩码
	subnetMask, err := getNetmask(h.options.NetMask)
	if err != nil {
		log.WithFields(h.sign).Errorf("Error parsing subnet mask %s", err.Error())
		return
	}

	router := parse(h.options.Router)
	dns := parse(h.options.DNS)

	// 构建 dhcp 响应包
	h.msg.UpdateOption(dhcpv4.OptMessageType(h.messageType))
	h.msg.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP(h.options.ServerIP)))
	h.msg.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseTime))
	h.msg.UpdateOption(dhcpv4.OptSubnetMask(subnetMask))
	h.msg.UpdateOption(dhcpv4.OptRouter(router...))
	h.msg.UpdateOption(dhcpv4.OptDNS(dns...))
	h.msg.BootFileName = h.options.BootFileName
	h.msg.YourIPAddr = assignedIP
	h.msg.ServerIPAddr = net.ParseIP(h.options.ServerIP)
	h.msg.GatewayIPAddr = net.ParseIP(h.options.GatewayIP)

	if _, err := h.conn.WriteTo(h.msg.ToBytes(), h.peer); err != nil {
		log.WithFields(h.sign).Errorf("Error Write DHCP reply message %s", err.Error())
	}
}

// 分配一个IP地址给客户端
func (h *Handler) createIP(rangeStart string, rangeEnd string) (net.IP, error) {
	var bind Binding
	var lease Leases

	// 检查这个客户端是否有绑定的IP地址
	if err := db.Where("client_hw_addr = ?", h.msg.ClientHWAddr.String()).First(&bind).Error; err == nil {
		// 如果 checkLeases 返回 true, 且 err 为 nil 则表示绑定的 IP 地址被分配了给其他机器
		if h.checkLeases(bind.BindAddr) {
			return nil, errors.New("the bound IP address is assigned to another machine")
		}
		return net.ParseIP(bind.BindAddr), nil
	}

	// 检查这个客户端是否已经分配了IP地址(如果已经分配则按照续约请求处理)
	if err := db.Where("client_hw_addr = ?", h.msg.ClientHWAddr.String()).First(&lease).Error; err == nil {
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
	return h.assignedIP(rangeStart, rangeEnd)
}

// 如果 addr 存在且 clientHW 相同则更新租约到期时间，并返回 false
// 如果 addr 存在且 clientHW 不同则返回 true ，表示此 ip 地址已经被分配
// 如果 addr 不存在则表示此 ip 地址尚未被分配，将租约信息写入到数据库，并返回 false
func (h *Handler) checkLeases(addr string) bool {
	var lease Leases

	options := QueryOptions()

	leaseTime, err := time.ParseDuration(options.LeaseTime)
	if err != nil {
		log.WithFields(h.sign).Errorf("Error lease generation time %s", err.Error())
		return true
	}

	// addr 存在且 clientHW 相同
	if err := db.Where("assigned_addr = ? and client_hw_addr = ?", addr, h.msg.ClientHWAddr.String()).First(&lease).Error; err == nil {
		lease.Expires = time.Now().Add(leaseTime)
		if err := db.Save(&lease).Error; err != nil {
			log.WithFields(h.sign).Errorf("Error update lease expires %s", err.Error())
			return true
		}
		return false
	}

	// addr 存在且 clientHW 不同，且expires值小于当前时间（表示此地址已经被分配给别的主机）
	if err := db.Where("assigned_addr = ?", addr, time.Now().Unix()).First(&lease).Error; err == nil {
		return true
	}

	lease.Expires = time.Now().Add(leaseTime)
	lease.AssignedAddr = addr
	lease.ClientHWAddr = h.msg.ClientHWAddr.String()
	if err := db.Create(&lease).Error; err != nil {
		log.WithFields(h.sign).Errorf("Error create lease info %s", err.Error())
		return true
	}
	return false
}

// 检查IP是否已被分配, 返回true表示已分配
func (h *Handler) checkIfTaken(ip net.IP) bool {
	var bind Binding
	var reserve Reserves
	addr := ip.String()
	if err := db.Where("bind_addr = ?", addr).First(&bind).Error; err == nil {
		return true
	}

	if err := db.Where("address = ?", addr).First(&reserve).Error; err == nil {
		return true
	}

	return h.checkLeases(addr)
}

// 从可分配的IP地址返回随机获取一个可用的IP地址
func (h *Handler) assignedIP(rangeStart string, rangeEnd string) (net.IP, error) {
	ip := make([]byte, 4)
	start := net.ParseIP(rangeStart)
	end := net.ParseIP(rangeEnd)
	rangeStartInt := binary.BigEndian.Uint32(start.To4())
	rangeEndInt := binary.BigEndian.Uint32(end.To4())
	binary.BigEndian.PutUint32(ip, random(rangeStartInt, rangeEndInt))
	taken := h.checkIfTaken(ip)
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt++
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt > rangeEndInt {
			break
		}
		taken = h.checkIfTaken(ip)
	}
	for taken {
		ipInt := binary.BigEndian.Uint32(ip)
		ipInt--
		binary.BigEndian.PutUint32(ip, ipInt)
		if ipInt < rangeStartInt {
			return nil, errors.New("no new ip addresses available")
		}
		taken = h.checkIfTaken(ip)
	}
	return ip, nil
}

func (h *Handler) ReleaseHandler() {
	h.withReleaseAddress("ReleaseHandler")
}

func (h *Handler) DeclineHandler() {
	h.withReleaseAddress("DeclineHandler")
}

func (h *Handler) withReleaseAddress(handlerName string) {
	var leases Leases
	clientHWAddr := h.msg.ClientHWAddr.String()
	if err := db.Unscoped().Where("client_hw_addr = ?", clientHWAddr).Delete(&leases).Error; err != nil {
		log.WithFields(h.sign).Warningf("%s release address %s", handlerName, err.Error())
	}
}
