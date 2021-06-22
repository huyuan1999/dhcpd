package main

import (
	"time"
)

// 配置信息(在数据库中应该也必须只能有一条配置信息存在)
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
	ACL          bool
	ACLAction    string // allow or deny
}

// 租约信息
type Leases struct {
	ClientHWAddr  string `gorm:"primarykey"`
	AssignedAddr  string `gorm:"unique"`
	Expires       time.Time
}

// 允许或者拒绝的客户端
type ACL struct {
	ClientHWAddr string
	Action       string
}

// mac 地址绑定
type Binding struct {
	ClientHWAddr string `gorm:"primarykey"`
	BindAddr     string `gorm:"unique"`
}

// 地址保留
type Reserves struct {
	Address string `gorm:"primarykey"`
}
