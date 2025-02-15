package db

import (
	"server/controller/db/mysql"
	"server/controller/recorder/common"
)

type Pod struct {
	OperatorBase[mysql.Pod]
}

func NewPod() *Pod {
	operater := &Pod{
		OperatorBase[mysql.Pod]{
			resourceTypeName: common.RESOURCE_TYPE_POD_EN,
			softDelete:       true,
		},
	}
	operater.setter = operater
	return operater
}

func (a *Pod) setDBItemID(dbItem *mysql.Pod, id int) {
	dbItem.ID = id
}
