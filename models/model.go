package models

import (
	"time"
)

// 配置信息(在数据库中应该也必须只能有一条配置信息存在)
type Options struct {
	LeaseTime    string `gorm:"unique" json:"lease_time" form:"lease_time" binding:"required"`
	ServerIP     string `gorm:"primarykey" json:"server_ip" form:"server_ip" binding:"required"`
	BootFileName string `gorm:"unique" json:"boot_file_name" form:"boot_file_name" binding:"required"`
	GatewayIP    string `gorm:"unique" json:"gateway_ip" form:"gateway_ip" binding:"required"`
	RangeStartIP string `gorm:"unique" json:"range_start_ip" form:"range_start_ip" binding:"required"`
	RangeEndIP   string `gorm:"unique" json:"range_end_ip" form:"range_end_ip" binding:"required"`
	NetMask      string `gorm:"unique" json:"net_mask" form:"net_mask" binding:"required"`
	Router       string `gorm:"unique" json:"router" form:"router" binding:"required"`
	DNS          string `gorm:"unique" json:"dns" form:"dns" binding:"required"`
	ACL          bool   `gorm:"unique" json:"acl" form:"acl" binding:"required"`
	// 当 ACLAction 为 allow 时默认的动作为 deny, 只有被匹配到的客户端才会分配地址
	// 当 ACLAction 为  deny 时默认的动作为 allow, 只有被匹配到的客户端才会被拒绝
	// allow or deny
	ACLAction string `gorm:"unique" json:"acl_action" form:"acl_action"`
}

// 租约信息
type Leases struct {
	ClientHWAddr string    `gorm:"primarykey" json:"client_hw_addr"`
	AssignedAddr string    `gorm:"unique" json:"assigned_addr"`
	Expires      time.Time `gorm:"not null" json:"expires"`
}

// 允许或者拒绝的客户端
type ACL struct {
	ClientHWAddr string `gorm:"primarykey" json:"client_hw_addr"`
	Action       string `gorm:"not null" json:"action"`
}

// mac 地址绑定
type Binding struct {
	ClientHWAddr string `gorm:"primarykey" json:"client_hw_addr"`
	BindAddr     string `gorm:"unique" json:"bind_addr"`
}

// 地址保留
type Reserves struct {
	Address string `gorm:"primarykey" json:"address"`
}
