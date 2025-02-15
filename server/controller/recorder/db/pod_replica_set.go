package db

import (
	"server/controller/db/mysql"
	"server/controller/recorder/common"
)

type PodReplicaSet struct {
	OperatorBase[mysql.PodReplicaSet]
}

func NewPodReplicaSet() *PodReplicaSet {
	operater := &PodReplicaSet{
		OperatorBase[mysql.PodReplicaSet]{
			resourceTypeName: common.RESOURCE_TYPE_POD_REPLICA_SET_EN,
			softDelete:       true,
		},
	}
	operater.setter = operater
	return operater
}

func (a *PodReplicaSet) setDBItemID(dbItem *mysql.PodReplicaSet, id int) {
	dbItem.ID = id
}
