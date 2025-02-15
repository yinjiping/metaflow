package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type LBListener struct {
	UpdaterBase[cloudmodel.LBListener, mysql.LBListener, *cache.LBListener]
}

func NewLBListener(wholeCache *cache.Cache, cloudData []cloudmodel.LBListener) *LBListener {
	updater := &LBListener{
		UpdaterBase[cloudmodel.LBListener, mysql.LBListener, *cache.LBListener]{
			cache:        wholeCache,
			dbOperator:   db.NewLBListener(),
			diffBaseData: wholeCache.LBListeners,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (l *LBListener) getDiffBaseByCloudItem(cloudItem *cloudmodel.LBListener) (diffBase *cache.LBListener, exists bool) {
	diffBase, exists = l.diffBaseData[cloudItem.Lcuuid]
	return
}

func (l *LBListener) generateDBItemToAdd(cloudItem *cloudmodel.LBListener) (*mysql.LBListener, bool) {
	lbID, exists := l.cache.ToolDataSet.GetLBIDByLcuuid(cloudItem.LBLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_LB_EN, cloudItem.LBLcuuid,
			common.RESOURCE_TYPE_LB_LISTENER_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysql.LBListener{
		Name:     cloudItem.Name,
		LBID:     lbID,
		IPs:      cloudItem.IPs,
		SNATIPs:  cloudItem.SNATIPs,
		Label:    cloudItem.Label,
		Port:     cloudItem.Port,
		Protocol: cloudItem.Protocol,
		Domain:   l.cache.DomainLcuuid,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

func (l *LBListener) generateUpdateInfo(diffBase *cache.LBListener, cloudItem *cloudmodel.LBListener) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if diffBase.Name != cloudItem.Name {
		updateInfo["name"] = cloudItem.Name
	}
	if diffBase.IPs != cloudItem.IPs {
		updateInfo["ips"] = cloudItem.IPs
	}
	if diffBase.SNATIPs != cloudItem.SNATIPs {
		updateInfo["snat_ips"] = cloudItem.SNATIPs
	}
	if diffBase.Port != cloudItem.Port {
		updateInfo["port"] = cloudItem.Port
	}
	if diffBase.Protocol != cloudItem.Protocol {
		updateInfo["protocol"] = cloudItem.Protocol
	}

	if len(updateInfo) > 0 {
		return updateInfo, true
	}
	return nil, false
}

func (l *LBListener) addCache(dbItems []*mysql.LBListener) {
	l.cache.AddLBListeners(dbItems)
}

func (l *LBListener) updateCache(cloudItem *cloudmodel.LBListener, diffBase *cache.LBListener) {
	diffBase.Update(cloudItem)
}

func (l *LBListener) deleteCache(lcuuids []string) {
	l.cache.DeleteLBListeners(lcuuids)
}
