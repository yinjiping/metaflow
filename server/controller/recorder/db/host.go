package db

import (
	"server/controller/db/mysql"
	"server/controller/recorder/common"
)

type Host struct {
	OperatorBase[mysql.Host]
}

func NewHost() *Host {
	operater := &Host{
		OperatorBase[mysql.Host]{
			resourceTypeName: common.RESOURCE_TYPE_HOST_EN,
			softDelete:       true,
		},
	}
	operater.setter = operater
	return operater
}

func (a *Host) setDBItemID(dbItem *mysql.Host, id int) {
	dbItem.ID = id
}
