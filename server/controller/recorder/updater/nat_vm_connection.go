package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type NATVMConnection struct {
	UpdaterBase[cloudmodel.NATVMConnection, mysql.NATVMConnection, *cache.NATVMConnection]
}

func NewNATVMConnection(wholeCache *cache.Cache, cloudData []cloudmodel.NATVMConnection) *NATVMConnection {
	updater := &NATVMConnection{
		UpdaterBase[cloudmodel.NATVMConnection, mysql.NATVMConnection, *cache.NATVMConnection]{
			cache:        wholeCache,
			dbOperator:   db.NewNATVMConnection(),
			diffBaseData: wholeCache.NATVMConnections,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (c *NATVMConnection) getDiffBaseByCloudItem(cloudItem *cloudmodel.NATVMConnection) (diffBase *cache.NATVMConnection, exists bool) {
	diffBase, exists = c.diffBaseData[cloudItem.Lcuuid]
	return
}

func (c *NATVMConnection) generateDBItemToAdd(cloudItem *cloudmodel.NATVMConnection) (*mysql.NATVMConnection, bool) {
	vmID, exists := c.cache.GetVMIDByLcuuid(cloudItem.VMLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_VM_EN, cloudItem.VMLcuuid,
			common.RESOURCE_TYPE_NAT_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	natID, exists := c.cache.GetNATGatewayIDByLcuuid(cloudItem.NATGatewayLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_NAT_GATEWAY_EN, cloudItem.NATGatewayLcuuid,
			common.RESOURCE_TYPE_NAT_VM_CONNECTION_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysql.NATVMConnection{
		Domain:       c.cache.DomainLcuuid,
		VMID:         vmID,
		NATGatewayID: natID,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

// 保留接口
func (c *NATVMConnection) generateUpdateInfo(diffBase *cache.NATVMConnection, cloudItem *cloudmodel.NATVMConnection) (map[string]interface{}, bool) {
	return nil, false
}

func (c *NATVMConnection) addCache(dbItems []*mysql.NATVMConnection) {
	c.cache.AddNATVMConnections(dbItems)
}

// 保留接口
func (c *NATVMConnection) updateCache(cloudItem *cloudmodel.NATVMConnection, diffBase *cache.NATVMConnection) {
}

func (c *NATVMConnection) deleteCache(lcuuids []string) {
	c.cache.DeleteNATVMConnections(lcuuids)
}
