package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type LANIP struct {
	UpdaterBase[cloudmodel.IP, mysql.LANIP, *cache.LANIP]
}

func NewLANIP(wholeCache *cache.Cache, cloudData []cloudmodel.IP) *LANIP {
	updater := &LANIP{
		UpdaterBase[cloudmodel.IP, mysql.LANIP, *cache.LANIP]{
			cache:        wholeCache,
			dbOperator:   db.NewLANIP(),
			diffBaseData: wholeCache.LANIPs,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (l *LANIP) getDiffBaseByCloudItem(cloudItem *cloudmodel.IP) (diffBase *cache.LANIP, exists bool) {
	diffBase, exists = l.diffBaseData[cloudItem.Lcuuid]
	return
}

func (l *LANIP) generateDBItemToAdd(cloudItem *cloudmodel.IP) (*mysql.LANIP, bool) {
	vinterfaceID, exists := l.cache.GetVInterfaceIDByLcuuid(cloudItem.VInterfaceLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_VINTERFACE_EN, cloudItem.VInterfaceLcuuid,
			common.RESOURCE_TYPE_LAN_IP_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	networkID, exists := l.cache.GetNetworkIDByVInterfaceLcuuid(cloudItem.VInterfaceLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_VINTERFACE_EN, cloudItem.VInterfaceLcuuid,
			common.RESOURCE_TYPE_LAN_IP_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	// netIndex, exists := l.cache.GetSubnetIndexByLcuuid(cloudItem.SubnetLcuuid)
	// if !exists {
	// 	log.Error(resourceAForResourceBNotFound(
	// 		common.RESOURCE_TYPE_SUBNET_EN, cloudItem.SubnetLcuuid,
	// 		common.RESOURCE_TYPE_LAN_IP_EN, cloudItem.Lcuuid,
	// 	))
	// 	return nil, false
	// }

	dbItem := &mysql.LANIP{
		IP:           common.FormatIP(cloudItem.IP),
		Domain:       l.cache.DomainLcuuid,
		SubDomain:    cloudItem.SubDomainLcuuid,
		NetworkID:    networkID,
		VInterfaceID: vinterfaceID,
		// NetIndex:     netIndex,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

// 保留接口
func (l *LANIP) generateUpdateInfo(diffBase *cache.LANIP, cloudItem *cloudmodel.IP) (map[string]interface{}, bool) {
	return nil, false
}

func (l *LANIP) addCache(dbItems []*mysql.LANIP) {
	l.cache.AddLANIPs(dbItems)
}

// 保留接口
func (l *LANIP) updateCache(cloudItem *cloudmodel.IP, diffBase *cache.LANIP) {
}

func (l *LANIP) deleteCache(lcuuids []string) {
	l.cache.DeleteLANIPs(lcuuids)
}
