package restful

import (
	"dhcp/server"
)

func aclReply(resMsg *ResMsg) {
	var acl []server.ACL
	options := server.QueryOptions()
	if options.ACL {
		if err := server.Db.Where("action = ?", options.ACLAction).Find(&acl).Error; err != nil {
			resMsg.Error = err.Error()
		}
		resMsg.Success = true
		resMsg.Data = acl
	}
}

func leasesReply(resMsg *ResMsg) () {
	var leases []server.Leases
	if err := server.Db.Find(&leases).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = leases
}

func bindReply(resMsg *ResMsg) {
	var bind []server.Binding
	if err := server.Db.Find(&bind).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = bind
}

func reserveReply(resMsg *ResMsg) () {
	var reserves []server.Reserves
	if err := server.Db.Find(&reserves).Error; err != nil {
		resMsg.Error = err.Error()
	}
	resMsg.Success = true
	resMsg.Data = reserves
}

