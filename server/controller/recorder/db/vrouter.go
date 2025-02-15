package db

import (
	"server/controller/db/mysql"
	"server/controller/recorder/common"
)

type VRouter struct {
	OperatorBase[mysql.VRouter]
}

func NewVRouter() *VRouter {
	operater := &VRouter{
		OperatorBase[mysql.VRouter]{
			resourceTypeName: common.RESOURCE_TYPE_VROUTER_EN,
			softDelete:       true,
		},
	}
	operater.setter = operater
	return operater
}

func (a *VRouter) setDBItemID(dbItem *mysql.VRouter, id int) {
	dbItem.ID = id
}
