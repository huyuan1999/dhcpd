package api

import (
	"dhcp/models"
	"dhcp/server"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"net/http"
)

type ResMsg struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Error   interface{} `json:"error"`
	Data    interface{} `json:"data"`
}

func verifyShouldBindJSON(c *gin.Context, obj interface{}) bool {
	var resMsg ResMsg
	if err := c.ShouldBindJSON(&obj); err != nil {
		resMsg.Error = err.Error()
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}

func respSuccess(c *gin.Context, data interface{}) {
	var resMsg ResMsg
	resMsg.Success = true
	resMsg.Data = data
	c.JSON(http.StatusOK, resMsg)
}

func respError(c *gin.Context, err interface{}) {
	var resMsg ResMsg
	resMsg.Error = err
	c.JSON(http.StatusOK, resMsg)
}

// @Summary 查询当前 DHCPD 配置信息
// @Description 查询当前 DHCPD 配置信息
// @Produce  json
// @Accept json
// @Param tag path string true "配置项" Enums(options, leases, acl, bind, reserve)
// @Success 200 {object} ResMsg
// @Router /api/v1/inform/{tag} [get]
func inform(c *gin.Context) {
	var resMsg ResMsg
	tag := c.Param("tag")
	switch tag {
	case "options":
		resMsg.Success = true
		resMsg.Data = server.QueryOptions()
	case "leases":
		leasesReply(&resMsg)
	case "acl":
		aclReply(&resMsg)
	case "bind":
		bindReply(&resMsg)
	case "reserve":
		reserveReply(&resMsg)
	default:
		resMsg.Error = "unknown inform"
	}
	c.JSON(http.StatusOK, resMsg)
}

// @Summary 添加 dhcpd 核心配置
// @Description 添加 dhcpd 核心配置, 包括地址, 路由, DNS等的分配
// @Produce  json
// @Accept json
// @Param message body Options true "添加 dhcpd 核心配置"
// @Success 200 {object} ResMsg
// @Router /api/v1/set/options/ [post]
func setOptions(c *gin.Context) {
	var resMsg ResMsg
	var options models.Options
	if !verifyShouldBindJSON(c, &options) {
		return
	}

	if !verifyOptions(c, options, resMsg) {
		return
	}

	err := object.Db.First(&models.Options{}).Error
	switch err {
	case nil:
		eMsg := "Error server is configured, please submit update request instead of create request"
		respError(c, eMsg)
	case gorm.ErrRecordNotFound:
		if err := object.Db.Create(&options).Error; err != nil {
			eMsg := fmt.Sprintf("Error creating server configuration information %s", err.Error())
			respError(c, eMsg)
			return
		}
		respSuccess(c, "success")
	default:
		respError(c, err.Error())
	}
}

// @Summary 添加 mac 地址绑定
// @Description mac 地址绑定(已被分配的地址需要等待客户端释放之后才能绑定)
// @Produce  json
// @Accept json
// @Param message body Binding true "添加 mac 地址绑定"
// @Success 200 {object} ResMsg
// @Router /api/v1/set/bind/ [post]
func setBind(c *gin.Context) {
	var resMsg ResMsg
	var bind models.Binding
	if !verifyShouldBindJSON(c, &bind) {
		return
	}

	if !verifyBind(c, bind, resMsg) {
		return
	}

	if err := object.Db.Create(&bind).Error; err != nil {
		respError(c, err)
		return
	}

	respSuccess(c, "success")
}

// @Summary 添加 acl 规则
// @Description 添加 acl 规则(acl规则必须在options中打开acl设置才能生效)
// @Produce  json
// @Accept json
// @Param message body ACL true "添加 acl 规则"
// @Success 200 {object} ResMsg
// @Router /api/v1/set/acl/ [post]
func setACL(c *gin.Context) {
	var resMsg ResMsg
	var acl models.ACL

	if !verifyShouldBindJSON(c, &acl) {
		return
	}

	if !verifyACL(c, acl, resMsg) {
		return
	}

	if err := object.Db.Create(&acl).Error; err != nil {
		respError(c, err)
		return
	}

	respSuccess(c, "success")
}

// @Summary 添加保留地址
// @Description 添加保留地址(已被分配的地址需要等待客户端释放之后才能被设置为保留地址)
// @Produce  json
// @Accept json
// @Param message body Reserves true "添加保留地址"
// @Success 200 {object} ResMsg
// @Router /api/v1/set/reserve/ [post]
func setReserve(c *gin.Context) {
	var resMsg ResMsg
	var reserve models.Reserves
	if !verifyShouldBindJSON(c, &reserve) {
		return
	}

	if !verifyReserve(c, reserve, resMsg) {
		return
	}

	if err := object.Db.Create(&reserve).Error; err != nil {
		respError(c, err)
		return
	}

	respSuccess(c, "success")
}

// @Summary 修改 dhcpd 核心配置
// @Description 修改 dhcpd 核心配置, 包括地址, 路由, DNS等的分配
// @Produce  json
// @Accept json
// @Param message body Options true "修改 dhcpd 核心配置"
// @Success 200 {object} ResMsg
// @Router /api/v1/update/options/ [put]
func updateOptions(c *gin.Context) {
	var resMsg ResMsg
	var options models.Options
	if !verifyShouldBindJSON(c, &options) {
		return
	}

	if !verifyOptions(c, options, resMsg) {
		return
	}

	if err := object.Db.Save(&options).Error; err != nil {
		respError(c, err.Error())
		return
	}

	respSuccess(c, "success")
}

// @Summary 修改 mac 地址绑定
// @Description mac 地址绑定(已被分配的地址需要等待客户端释放之后才能绑定)
// @Produce  json
// @Accept json
// @Param message body Binding true "修改 mac 地址绑定"
// @Success 200 {object} ResMsg
// @Router /api/v1/update/bind/ [put]
func updateBind(c *gin.Context) {
	var resMsg ResMsg
	var bind models.Binding
	if !verifyShouldBindJSON(c, &bind) {
		return
	}

	if !verifyBind(c, bind, resMsg) {
		return
	}

	if err := object.Db.Save(&bind).Error; err != nil {
		respError(c, err.Error())
		return
	}
	respSuccess(c, "success")
}

// @Summary 修改 acl 规则
// @Description 修改 acl 规则(acl规则必须在options中打开acl设置才能生效)
// @Produce  json
// @Accept json
// @Param message body ACL true "修改 acl 规则"
// @Success 200 {object} ResMsg
// @Router /api/v1/update/acl/ [put]
func updateACL(c *gin.Context) {
	var resMsg ResMsg
	var acl models.ACL
	if !verifyShouldBindJSON(c, &acl) {
		return
	}

	if !verifyACL(c, acl, resMsg) {
		return
	}

	if err := object.Db.Save(&acl).Error; err != nil {
		respError(c, err.Error())
		return
	}
	respSuccess(c, "success")
}

// @Summary 删除匹配的 mac 地址绑定规则
// @Description 删除匹配的 mac 地址绑定规则
// @Produce  json
// @Accept json
// @Param mac query string false "通过 mac 地址匹配需要删除的绑定规则(mac 或者 ip 至少指定一项)"
// @Param ip query string false "通过 ip 地址匹配需要删除的绑定规则(mac 或者 ip 至少指定一项)"
// @Success 200 {object} ResMsg
// @Router /api/v1/del/bind/ [delete]
func deleteBind(c *gin.Context) {
	mac := c.Request.FormValue("mac")
	ip := c.Request.FormValue("ip")

	_, err := net.ParseMAC(mac)
	if net.ParseIP(ip) == nil && err != nil {
		respError(c, "please specify a valid mac or ip")
		return
	}

	if net.ParseIP(ip) != nil && err == nil {
		if err := object.Db.Unscoped().Where("client_hw_addr = ? and bind_addr = ?", mac, ip).Delete(&models.Binding{}).Error; err != nil {
			respError(c, err)
			return
		}
		respSuccess(c, "success")
		return
	}

	if net.ParseIP(ip) != nil {
		if err := object.Db.Unscoped().Where("bind_addr = ?", ip).Delete(&models.Binding{}).Error; err != nil {
			respError(c, err)
			return
		}
		respSuccess(c, "success")
		return
	}

	if err := object.Db.Unscoped().Where("client_hw_addr = ?", mac).Delete(&models.Binding{}).Error; err != nil {
		respError(c, err)
		return
	}
	respSuccess(c, "success")
	return
}

// @Summary 删除匹配的 acl 规则
// @Description 删除匹配的 acl 规则
// @Produce  json
// @Accept json
// @Param mac query string false "通过 mac 地址匹配需要删除的 acl 规则"
// @Success 200 {object} ResMsg
// @Router /api/v1/del/acl/ [delete]
func deleteACL(c *gin.Context) {
	mac := c.Request.FormValue("mac")
	if _, err := net.ParseMAC(mac); err != nil {
		respError(c, err)
		return
	}

	if err := object.Db.Unscoped().Where("client_hw_addr = ?", mac).Delete(&models.ACL{}).Error; err != nil {
		respError(c, err)
		return
	}
	respSuccess(c, "success")
}

// @Summary 删除保留 IP
// @Description 删除保留 IP
// @Produce  json
// @Accept json
// @Param ip query string false "通过 ip 地址删除保留 IP"
// @Success 200 {object} ResMsg
// @Router /api/v1/del/reserve/ [delete]
func deleteReserve(c *gin.Context) {
	ip := c.Request.FormValue("ip")
	if net.ParseIP(ip) == nil {
		respError(c, "invalid ip address")
		return
	}

	if err := object.Db.Unscoped().Where("address = ?", ip).Delete(&models.Reserves{}).Error; err != nil {
		respError(c, err)
		return
	}
	respSuccess(c, "success")
}
