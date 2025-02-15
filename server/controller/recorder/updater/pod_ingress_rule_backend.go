package updater

import (
	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
	"server/controller/recorder/common"
	"server/controller/recorder/db"
)

type PodIngressRuleBackend struct {
	UpdaterBase[cloudmodel.PodIngressRuleBackend, mysql.PodIngressRuleBackend, *cache.PodIngressRuleBackend]
}

func NewPodIngressRuleBackend(wholeCache *cache.Cache, cloudData []cloudmodel.PodIngressRuleBackend) *PodIngressRuleBackend {
	updater := &PodIngressRuleBackend{
		UpdaterBase[cloudmodel.PodIngressRuleBackend, mysql.PodIngressRuleBackend, *cache.PodIngressRuleBackend]{
			cache:        wholeCache,
			dbOperator:   db.NewPodIngressRuleBackend(),
			diffBaseData: wholeCache.PodIngressRuleBackends,
			cloudData:    cloudData,
		},
	}
	updater.dataGenerator = updater
	updater.cacheHandler = updater
	return updater
}

func (b *PodIngressRuleBackend) getDiffBaseByCloudItem(cloudItem *cloudmodel.PodIngressRuleBackend) (diffBase *cache.PodIngressRuleBackend, exists bool) {
	diffBase, exists = b.diffBaseData[cloudItem.Lcuuid]
	return
}

func (b *PodIngressRuleBackend) generateDBItemToAdd(cloudItem *cloudmodel.PodIngressRuleBackend) (*mysql.PodIngressRuleBackend, bool) {
	podIngressRuleID, exists := b.cache.ToolDataSet.GetPodIngressRuleIDByLcuuid(cloudItem.PodIngressRuleLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_POD_INGRESS_RULE_EN, cloudItem.PodIngressRuleLcuuid,
			common.RESOURCE_TYPE_POD_INGRESS_RULE_BACKEND_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	podIngressID, exists := b.cache.ToolDataSet.GetPodIngressIDByLcuuid(cloudItem.PodIngressLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_POD_INGRESS_EN, cloudItem.PodIngressLcuuid,
			common.RESOURCE_TYPE_POD_INGRESS_RULE_BACKEND_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}
	podServiceID, exists := b.cache.ToolDataSet.GetPodServiceIDByLcuuid(cloudItem.PodServiceLcuuid)
	if !exists {
		log.Errorf(resourceAForResourceBNotFound(
			common.RESOURCE_TYPE_POD_SERVICE_EN, cloudItem.PodServiceLcuuid,
			common.RESOURCE_TYPE_POD_INGRESS_RULE_BACKEND_EN, cloudItem.Lcuuid,
		))
		return nil, false
	}

	dbItem := &mysql.PodIngressRuleBackend{
		Path:             cloudItem.Path,
		Port:             cloudItem.Port,
		PodServiceID:     podServiceID,
		PodIngressID:     podIngressID,
		PodIngressRuleID: podIngressRuleID,
		SubDomain:        cloudItem.SubDomainLcuuid,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

// 保留接口
func (b *PodIngressRuleBackend) generateUpdateInfo(diffBase *cache.PodIngressRuleBackend, cloudItem *cloudmodel.PodIngressRuleBackend) (map[string]interface{}, bool) {
	return nil, false
}

func (b *PodIngressRuleBackend) addCache(dbItems []*mysql.PodIngressRuleBackend) {
	b.cache.AddPodIngressRuleBackends(dbItems)
}

// 保留接口
func (b *PodIngressRuleBackend) updateCache(cloudItem *cloudmodel.PodIngressRuleBackend, diffBase *cache.PodIngressRuleBackend) {
}

func (b *PodIngressRuleBackend) deleteCache(lcuuids []string) {
	b.cache.DeletePodIngressRuleBackends(lcuuids)
}
