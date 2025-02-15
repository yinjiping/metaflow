package updater

import (
	"strings"

	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type WANIP struct {
	UpdaterBase[cloudmodel.IP, mysql.WANIP, *cache.WANIP]
}

func NewWANIP(wholeCache *cache.Cache, cloudData []cloudmodel.IP) *WANIP {
	updater := &WANIP{
		UpdaterBase[cloudmodel.IP, mysql.WANIP, *cache.WANIP]{
			cache:        wholeCache,
			dbOperator:   db.NewWANIP(),
			diffBaseData: wholeCache.WANIPs,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (l *WANIP) getDiffBaseByCloudItem(cloudItem *cloudmodel.IP) (diffBase *cache.WANIP, exists bool) {
	diffBase, exists = l.diffBaseData[cloudItem.Lcuuid]
	return
}

func (l *WANIP) generateDBItemToAdd(cloudItem *cloudmodel.IP) (*mysql.WANIP, bool) {
	vinterfaceID, exists := l.cache.GetVInterfaceIDByLcuuid(cloudItem.VInterfaceLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_VINTERFACE_EN, cloudItem.VInterfaceLcuuid,
			common.RESOURCE_TYPE_WAN_IP_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysql.WANIP{
		IP:           common.FormatIP(cloudItem.IP),
		Domain:       l.cache.DomainLcuuid,
		SubDomain:    cloudItem.SubDomainLcuuid,
		VInterfaceID: vinterfaceID,
		Region:       cloudItem.RegionLcuuid,
		ISP:          common.WAN_IP_ISP,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	if strings.Contains(cloudItem.IP, ":") {
		dbItem.Netmask = common.IPV6_DEFAULT_NETMASK
		dbItem.Gateway = common.IPV6_DEFAULT_GATEWAY
	} else {
		dbItem.Netmask = common.IPV4_DEFAULT_NETMASK
		dbItem.Gateway = common.IPV4_DEFAULT_GATEWAY
	}
	return dbItem, true
}

func (l *WANIP) generateUpdateInfo(diffBase *cache.WANIP, cloudItem *cloudmodel.IP) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if diffBase.RegionLcuuid != cloudItem.RegionLcuuid {
		updateInfo["region"] = cloudItem.RegionLcuuid
	}
	return updateInfo, len(updateInfo) > 0
}

func (l *WANIP) addCache(dbItems []*mysql.WANIP) {
	l.cache.AddWANIPs(dbItems)
}

func (l *WANIP) updateCache(cloudItem *cloudmodel.IP, diffBase *cache.WANIP) {
	diffBase.Update(cloudItem)
}

func (l *WANIP) deleteCache(lcuuids []string) {
	l.cache.DeleteWANIPs(lcuuids)
}
