package main

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"strings"
)

// 检查子网掩码是否合法
func checkValidNetmask(netmask net.IPMask) bool {
	netmaskInt := binary.BigEndian.Uint32(netmask)
	x := ^netmaskInt
	y := x + 1
	return (y & x) == 0
}

// 将合法的子网掩码转换为 net.IPMask 对象
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
