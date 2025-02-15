package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type PodGroup struct {
	UpdaterBase[cloudmodel.PodGroup, mysql.PodGroup, *cache.PodGroup]
}

func NewPodGroup(wholeCache *cache.Cache, cloudData []cloudmodel.PodGroup) *PodGroup {
	updater := &PodGroup{
		UpdaterBase[cloudmodel.PodGroup, mysql.PodGroup, *cache.PodGroup]{
			cache:        wholeCache,
			dbOperator:   db.NewPodGroup(),
			diffBaseData: wholeCache.PodGroups,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (p *PodGroup) getDiffBaseByCloudItem(cloudItem *cloudmodel.PodGroup) (diffBase *cache.PodGroup, exists bool) {
	diffBase, exists = p.diffBaseData[cloudItem.Lcuuid]
	return
}

func (p *PodGroup) generateDBItemToAdd(cloudItem *cloudmodel.PodGroup) (*mysql.PodGroup, bool) {
	podNamespaceID, exists := p.cache.ToolDataSet.GetPodNamespaceIDByLcuuid(cloudItem.PodNamespaceLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_POD_NAMESPACE_EN, cloudItem.PodNamespaceLcuuid,
			common.RESOURCE_TYPE_POD_GROUP_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	podClusterID, exists := p.cache.ToolDataSet.GetPodClusterIDByLcuuid(cloudItem.PodClusterLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_POD_CLUSTER_EN, cloudItem.PodClusterLcuuid,
			common.RESOURCE_TYPE_POD_GROUP_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	dbItem := &mysql.PodGroup{
		Name:           cloudItem.Name,
		Type:           cloudItem.Type,
		Label:          cloudItem.Label,
		PodNum:         cloudItem.PodNum,
		PodNamespaceID: podNamespaceID,
		PodClusterID:   podClusterID,
		SubDomain:      cloudItem.SubDomainLcuuid,
		Domain:         p.cache.DomainLcuuid,
		Region:         cloudItem.RegionLcuuid,
		AZ:             cloudItem.AZLcuuid,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

func (p *PodGroup) generateUpdateInfo(diffBase *cache.PodGroup, cloudItem *cloudmodel.PodGroup) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if diffBase.Name != cloudItem.Name {
		updateInfo["name"] = cloudItem.Name
	}
	if diffBase.Type != cloudItem.Type {
		updateInfo["type"] = cloudItem.Type
	}
	if diffBase.PodNum != cloudItem.PodNum {
		updateInfo["pod_num"] = cloudItem.PodNum
	}
	if diffBase.Label != cloudItem.Label {
		updateInfo["label"] = cloudItem.Label
	}
	if diffBase.RegionLcuuid != cloudItem.RegionLcuuid {
		updateInfo["region"] = cloudItem.RegionLcuuid
	}
	if diffBase.AZLcuuid != cloudItem.AZLcuuid {
		updateInfo["az"] = cloudItem.AZLcuuid
	}
	if len(updateInfo) > 0 {
		return updateInfo, true
	}
	return nil, false
}

func (p *PodGroup) addCache(dbItems []*mysql.PodGroup) {
	p.cache.AddPodGroups(dbItems)
}

func (p *PodGroup) updateCache(cloudItem *cloudmodel.PodGroup, diffBase *cache.PodGroup) {
	diffBase.Update(cloudItem)
}

func (p *PodGroup) deleteCache(lcuuids []string) {
	p.cache.DeletePodGroups(lcuuids)
}
