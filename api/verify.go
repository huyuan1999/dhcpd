package api

import (
	"dhcp/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"net/http"
)

func verifyOptions(c *gin.Context, options models.Options, resMsg ResMsg) bool {
	if options.ACL && !(options.ACLAction == "allow" || options.ACLAction == "deny") {
		resMsg.Error = "Error enable acl without specifying acl action (allow|deny)"
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}

func verifyBind(c *gin.Context, bind models.Binding, resMsg ResMsg) bool {
	// ip 和 mac 是否合法
	_, err := net.ParseMAC(bind.ClientHWAddr)
	if net.ParseIP(bind.BindAddr) == nil || err != nil {
		resMsg.Error = "invalid mac address or invalid bind address"
		c.JSON(http.StatusOK, resMsg)
		return false
	}

	// 是否已被分配
	if err := object.Db.Where("assigned_addr = ?", bind.BindAddr).First(&models.Leases{}).Error; err != gorm.ErrRecordNotFound {
		resMsg.Error = "bind address assigned"
		c.JSON(http.StatusOK, resMsg)
		return false
	}

	// 是否是保留地址
	if err := object.Db.Where("address = ?", bind.BindAddr).First(&models.Reserves{}).Error; err != gorm.ErrRecordNotFound {
		resMsg.Error = "the binding address is a reserved address"
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}

func verifyACL(c *gin.Context, acl models.ACL, resMsg ResMsg) bool {
	if acl.Action != "allow" && acl.Action != "deny" {
		resMsg.Error = "action in (allow|deny)"
		c.JSON(http.StatusOK, resMsg)
		return false
	}

	_, err := net.ParseMAC(acl.ClientHWAddr)
	if err != nil {
		resMsg.Error = err.Error()
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}

func verifyReserve(c *gin.Context, reserve models.Reserves, resMsg ResMsg) bool {
	if net.ParseIP(reserve.Address) == nil {
		resMsg.Error = "invalid reserve address"
		c.JSON(http.StatusOK, resMsg)
		return false
	}

	if err := object.Db.Where("assigned_addr = ?", reserve.Address).First(&models.Leases{}).Error; err != gorm.ErrRecordNotFound {
		resMsg.Error = "bind address assigned"
		c.JSON(http.StatusOK, resMsg)
		return false
	}

	if err := object.Db.Where("bind_addr = ?", reserve.Address).First(&models.Binding{}).Error; err != gorm.ErrRecordNotFound {
		resMsg.Error = "the reserved address has been bound to the client"
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}
