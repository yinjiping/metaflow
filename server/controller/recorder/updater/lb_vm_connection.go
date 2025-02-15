package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type LBVMConnection struct {
	UpdaterBase[cloudmodel.LBVMConnection, mysql.LBVMConnection, *cache.LBVMConnection]
}

func NewLBVMConnection(wholeCache *cache.Cache, cloudData []cloudmodel.LBVMConnection) *LBVMConnection {
	updater := &LBVMConnection{
		UpdaterBase[cloudmodel.LBVMConnection, mysql.LBVMConnection, *cache.LBVMConnection]{
			cache:        wholeCache,
			dbOperator:   db.NewLBVMConnection(),
			diffBaseData: wholeCache.LBVMConnections,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (c *LBVMConnection) getDiffBaseByCloudItem(cloudItem *cloudmodel.LBVMConnection) (diffBase *cache.LBVMConnection, exists bool) {
	diffBase, exists = c.diffBaseData[cloudItem.Lcuuid]
	return
}

func (c *LBVMConnection) generateDBItemToAdd(cloudItem *cloudmodel.LBVMConnection) (*mysql.LBVMConnection, bool) {
	vmID, exists := c.cache.GetVMIDByLcuuid(cloudItem.VMLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_VM_EN, cloudItem.VMLcuuid,
			common.RESOURCE_TYPE_LB_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	lbID, exists := c.cache.GetLBIDByLcuuid(cloudItem.LBLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_LB_EN, cloudItem.LBLcuuid,
			common.RESOURCE_TYPE_LB_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysql.LBVMConnection{
		Domain: c.cache.DomainLcuuid,
		VMID:   vmID,
		LBID:   lbID,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

// 保留接口
func (c *LBVMConnection) generateUpdateInfo(diffBase *cache.LBVMConnection, cloudItem *cloudmodel.LBVMConnection) (map[string]interface{}, bool) {
	return nil, false
}

func (c *LBVMConnection) addCache(dbItems []*mysql.LBVMConnection) {
	c.cache.AddLBVMConnections(dbItems)
}

// 保留接口
func (c *LBVMConnection) updateCache(cloudItem *cloudmodel.LBVMConnection, diffBase *cache.LBVMConnection) {
}

func (c *LBVMConnection) deleteCache(lcuuids []string) {
	c.cache.DeleteLBVMConnections(lcuuids)
}
