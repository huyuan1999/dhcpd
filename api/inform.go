package api

import (
	"dhcp/models"
	"dhcp/server"
)

func aclReply(resMsg *ResMsg) {
	var acl []models.ACL
	options := server.QueryOptions()
	if options.ACL {
		if err := object.Db.Where("action = ?", options.ACLAction).Find(&acl).Error; err != nil {
			resMsg.Error = err.Error()
		}
		resMsg.Success = true
		resMsg.Data = acl
	}
}

func leasesReply(resMsg *ResMsg) () {
	var leases []models.Leases
	if err := object.Db.Find(&leases).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = leases
}

func bindReply(resMsg *ResMsg) {
	var bind []models.Binding
	if err := object.Db.Find(&bind).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = bind
}

func reserveReply(resMsg *ResMsg) () {
	var reserves []models.Reserves
	if err := object.Db.Find(&reserves).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = reserves
}
