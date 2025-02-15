package vtap

import (
	"fmt"
	"strings"
	"sync"

	. "server/controller/common"
	. "server/controller/trisolaris/common"
	"server/controller/trisolaris/metadata"
	. "server/controller/trisolaris/utils"
)

var ALL_DOMAIMS = []string{"0"}

type VTapPlatformData struct {

	// 下的云平台列表=xxx，容器集群内部IP下发=所有集群
	// key为vtap_group_lcuuid
	platformDataType1 *PlatformDataType

	// 下发的云平台列表=全部，容器集群内部IP下发=采集器所在集群
	// key为vtap_group_lcuuid+采集器所在容器集群LCUUID
	platformDataType2 *PlatformDataType

	// 下发的云平台列表=xxx，容器集群内部IP下发=采集器所在集群
	// key为vtap_group_lcuuid+集器所在容器集群LCUUID
	platformDataType3 *PlatformDataType

	// 专属采集器
	platformDataBMDedicated *PlatformDataType
}

func newVTapPlatformData() *VTapPlatformData {
	return &VTapPlatformData{
		platformDataType1:       newPlatformDataType("platformDataType1"),
		platformDataType2:       newPlatformDataType("platformDataType2"),
		platformDataType3:       newPlatformDataType("platformDataType3"),
		platformDataBMDedicated: newPlatformDataType("platformDataBMDedicated"),
	}
}

func (v *VTapPlatformData) String() string {
	log.Debug(v.platformDataType1)
	log.Debug(v.platformDataType2)
	log.Debug(v.platformDataType3)
	log.Debug(v.platformDataBMDedicated)
	return "vtap Platform data"
}

type PlatformDataType struct {
	sync.RWMutex
	platformDataMap map[string]*metadata.PlatformData
	name            string
}

func newPlatformDataType(name string) *PlatformDataType {
	return &PlatformDataType{
		platformDataMap: make(map[string]*metadata.PlatformData),
		name:            name,
	}
}

func (t *PlatformDataType) String() string {
	t.RLock()
	defer t.RUnlock()
	for k, v := range t.platformDataMap {
		log.Debugf("key: [%s]; value:[%s]", k, v)
	}
	return t.name
}

func (t *PlatformDataType) setPlatformDataCache(key string, data *metadata.PlatformData) {
	t.Lock()
	defer t.Unlock()
	t.platformDataMap[key] = data
}

func (t *PlatformDataType) getPlatformDataCache(key string) *metadata.PlatformData {
	t.RLock()
	defer t.RUnlock()
	return t.platformDataMap[key]
}

func (t *PlatformDataType) clearCache() {
	t.Lock()
	defer t.Unlock()
	t.platformDataMap = make(map[string]*metadata.PlatformData)
}

func (v *VTapPlatformData) clearPlatformDataTypeCache() {
	v.platformDataType1.clearCache()
	v.platformDataType2.clearCache()
	v.platformDataType3.clearCache()
	v.platformDataBMDedicated.clearCache()
}

func (v *VTapPlatformData) setPlatformDataByVTap(p *metadata.PlatformDataOP, c *VTapCache) {
	if c.GetVTapType() == VTAP_TYPE_KVM {
		v.setSkipPlatformDataByVTap(p, c)
	} else {
		v.setNormalPlatformDataByVTap(p, c)
	}
}

func (v *VTapPlatformData) setNormalPlatformDataByVTap(p *metadata.PlatformDataOP, c *VTapCache) {
	vTapType := c.GetVTapType()
	// 隧道解封装采集器没有平台数据
	if vTapType == VTAP_TYPE_TUNNEL_DECAPSULATION {
		return
	}

	log.Debug(c.GetCtrlIP())
	log.Debug(c.GetCtrlMac())
	log.Debug(c.getPodDomains())
	vTapGroupLcuuid := c.GetVTapGroupLcuuid()
	vtapConfig := c.GetVTapConfig()
	if vtapConfig == nil {
		return
	}
	log.Debug(vtapConfig.PodClusterInternalIP, vtapConfig.ConvertedDomains)
	if vtapConfig.PodClusterInternalIP == ALL_CLUSTERS &&
		SliceEqual[string](vtapConfig.ConvertedDomains, ALL_DOMAIMS) {
		// 下发的云平台列表=全部，容器集群内部IP下发=所有集群
		// 所有云平台所有数据

		log.Debug("all:", p.GetAllSimplePlatformData())
		c.setVTapPlatformData(p.GetAllSimplePlatformData())
	} else if vtapConfig.PodClusterInternalIP == ALL_CLUSTERS {
		// 下发的云平台列表=xxx，容器集群内部IP下发=所有集群
		// 云平台列表=xxx的所有数据

		// 获取缓存数据
		data := v.platformDataType1.getPlatformDataCache(vTapGroupLcuuid)
		if data != nil {
			c.setVTapPlatformData(data)
			return
		}
		domainToAllPlatformData := p.GetDomainToAllPlatformData()
		domainAllData := metadata.NewPlatformData("platformDataType1", "", 0, PLATFORM_DATA_TYPE_1)
		for _, domainLcuuid := range vtapConfig.ConvertedDomains {
			domainData := domainToAllPlatformData[domainLcuuid]
			if domainData == nil {
				log.Errorf("domain(%s) no platform data", domainLcuuid)
				continue
			}
			domainAllData.Merge(domainData)
		}
		domainAllData.GeneratePlatformDataResult()
		v.platformDataType1.setPlatformDataCache(vTapGroupLcuuid, domainAllData)
		c.setVTapPlatformData(domainAllData)
		log.Debug(domainAllData)
	} else if vtapConfig.PodClusterInternalIP == CLUSTER_OF_VTAP &&
		SliceEqual[string](vtapConfig.ConvertedDomains, ALL_DOMAIMS) {
		// 下发的云平台列表=全部，容器集群内部IP下发=采集器所在集群
		// 所有云平台中devicetype != POD/容器服务的所有接口，采集器所在集群devicetype=POD/容器服务的所有接口

		// 专属服务器类型：所有集群
		if vTapType == VTAP_TYPE_DEDICATED {
			data := p.GetAllSimplePlatformData()
			c.setVTapPlatformData(data)
			log.Debug("vtap_type_dedicated: ", data)
			return
		}
		// 获取缓存数据
		podDomains := c.getPodDomains()
		key := fmt.Sprintf("%s+%s", vTapGroupLcuuid, strings.Join(podDomains, "+"))
		data := v.platformDataType2.getPlatformDataCache(key)
		if data != nil {
			c.setVTapPlatformData(data)
			return
		}
		domainToPlarformDataOnlyPod := p.GetDomainToPlatformDataOnlyPod()
		domainAllData := metadata.NewPlatformData("platformDataType2", "", 0, PLATFORM_DATA_TYPE_2)
		domainAllData.Merge(p.GetAllSimplePlatformDataExceptPod())
		for _, podDomain := range podDomains {
			vTapDomainData := domainToPlarformDataOnlyPod[podDomain]
			if vTapDomainData == nil {
				log.Errorf("vtap pod domain(%s) no data", podDomain)
				continue
			}
			domainAllData.MergeInterfaces(vTapDomainData)
		}
		domainAllData.GeneratePlatformDataResult()
		c.setVTapPlatformData(domainAllData)
		v.platformDataType2.setPlatformDataCache(key, domainAllData)
		log.Debug(domainAllData)
	} else if vtapConfig.PodClusterInternalIP == CLUSTER_OF_VTAP {
		// 下发的云平台列表=xxx，容器集群内部IP下发=采集器所在集群
		// 云平台列表=xxx中devicetype != POD/容器服务所有接口，集器所在集群devicetype=POD/容器服务的所有接口

		// 专属服务器类型：下发的云平台列表=xxx，容器集群内部IP下发=所有集群
		if vTapType == VTAP_TYPE_DEDICATED {
			// 获取缓存数据
			data := v.platformDataBMDedicated.getPlatformDataCache(vTapGroupLcuuid)
			if data != nil {
				c.setVTapPlatformData(data)
				return
			}
			domainToAllPlatformData := p.GetDomainToAllPlatformData()
			domainAllData := metadata.NewPlatformData("platformDataBMDedicated", "", 0, PLATFORM_DATA_BM_DEDICATED)
			for _, domainLcuuid := range vtapConfig.ConvertedDomains {
				domainData := domainToAllPlatformData[domainLcuuid]
				if domainData == nil {
					log.Errorf("domain(%s) no platform data", domainLcuuid)
					continue
				}
				domainAllData.Merge(domainData)
			}
			domainAllData.GeneratePlatformDataResult()
			c.setVTapPlatformData(domainAllData)
			v.platformDataBMDedicated.setPlatformDataCache(vTapGroupLcuuid, domainAllData)
			log.Debug(domainAllData)
			return
		}

		// 获取缓存数据
		podDomains := c.getPodDomains()
		key := fmt.Sprintf("%s+%s", vTapGroupLcuuid, strings.Join(podDomains, "+"))
		data := v.platformDataType3.getPlatformDataCache(key)
		if data != nil {
			c.setVTapPlatformData(data)
			return
		}

		domainToPlatformDataExceptPod := p.GetDomainToPlatformDataExceptPod()
		domainAllData := metadata.NewPlatformData("platformDataType3", "", 0, PLATFORM_DATA_TYPE_3)
		for _, domainLcuuid := range vtapConfig.ConvertedDomains {
			domainData := domainToPlatformDataExceptPod[domainLcuuid]
			if domainData == nil {
				log.Errorf("domain(%s) no platform data", domainLcuuid)
				continue
			}
			domainAllData.Merge(domainData)
		}

		for _, podDomain := range podDomains {
			vtapDomainData := domainToPlatformDataExceptPod[podDomain]
			if vtapDomainData == nil {
				log.Errorf("domain(%s) no platform data", podDomain)
				continue
			}
			if Find[string](vtapConfig.ConvertedDomains, podDomain) {
				domainAllData.MergeInterfaces(vtapDomainData)
			} else {
				domainAllData.Merge(vtapDomainData)
			}
		}

		domainAllData.GeneratePlatformDataResult()
		c.setVTapPlatformData(domainAllData)
		v.platformDataType3.setPlatformDataCache(key, domainAllData)
		log.Debug(domainAllData)
	}
}

func (v *VTapPlatformData) setSkipPlatformDataByVTap(p *metadata.PlatformDataOP, c *VTapCache) {
	log.Debug(c.GetCtrlIP())
	log.Debug(c.GetCtrlMac())
	log.Debug(c.getPodDomains())
	vtapConfig := c.GetVTapConfig()
	if vtapConfig == nil {
		return
	}
	log.Debug(vtapConfig.PodClusterInternalIP, vtapConfig.ConvertedDomains)
	if vtapConfig.PodClusterInternalIP == ALL_CLUSTERS &&
		SliceEqual[string](vtapConfig.ConvertedDomains, ALL_DOMAIMS) {
		// 下发的云平台列表=全部，容器集群内部IP下发=所有集群
		// 所有云平台所有数据
		serverToData := p.GetServerToSkipAllSimplePlatformData()
		if data, ok := serverToData[c.GetLaunchServer()]; ok {
			c.setVTapPlatformData(data)
		} else {
			log.Debug("all:", p.GetAllSimplePlatformData())
			c.setVTapPlatformData(p.GetAllSimplePlatformData())
		}
	} else if vtapConfig.PodClusterInternalIP == ALL_CLUSTERS {
		// 下发的云平台列表=xxx，容器集群内部IP下发=所有集群
		// 云平台列表=xxx的所有数据

		domainToAllPlatformData := p.GetDomainToAllPlatformData()
		domainToSkipAllPlatformData := p.GetDomainToSkipAllPlatformData()
		domainAllData := metadata.NewPlatformData("skipPlatformDataType1", "", 0, SKIP_PLATFORM_DATA_TYPE_1)
		skipVifIDs := p.GetRawData().GetSkipVifIDs(c.GetLaunchServer())
		for _, domainLcuuid := range vtapConfig.ConvertedDomains {
			domainData, ok := domainToSkipAllPlatformData[domainLcuuid]
			if ok == true {
				domainAllData.SkipMerge(domainData, skipVifIDs)
				continue
			}
			domainData, ok = domainToAllPlatformData[domainLcuuid]
			if ok == true {
				domainAllData.Merge(domainData)
				continue
			}
			log.Errorf("domain(%s) no platform data", domainLcuuid)
		}
		domainAllData.GeneratePlatformDataResult()
		c.setVTapPlatformData(domainAllData)
		log.Debug(domainAllData)
	} else if vtapConfig.PodClusterInternalIP == CLUSTER_OF_VTAP &&
		SliceEqual[string](vtapConfig.ConvertedDomains, ALL_DOMAIMS) {
		// 下发的云平台列表=全部，容器集群内部IP下发=采集器所在集群
		// 所有云平台中devicetype != POD/容器服务的所有接口，采集器所在集群devicetype=POD/容器服务的所有接口

		skipVifIDs := p.GetRawData().GetSkipVifIDs(c.GetLaunchServer())
		podDomains := c.getPodDomains()
		domainToPlarformDataOnlyPod := p.GetDomainToPlatformDataOnlyPod()
		domainToSkipPlarformDataOnlyPod := p.GetDomainToSkipPlatformDataOnlyPod()
		domainAllData := metadata.NewPlatformData("skipPlatformDataType2", "", 0, SKIP_PLATFORM_DATA_TYPE_2)
		domainAllData.SkipMerge(p.GetSkipAllSimplePlatformDataExceptPod(), skipVifIDs)
		for _, podDomain := range podDomains {
			if vTapDomainData, ok := domainToSkipPlarformDataOnlyPod[podDomain]; ok {
				domainAllData.SkipMergeInterfaces(vTapDomainData, skipVifIDs)
				continue
			}
			if vTapDomainData, ok := domainToPlarformDataOnlyPod[podDomain]; ok {
				domainAllData.MergeInterfaces(vTapDomainData)
				continue
			}
			log.Errorf("vtap pod domain(%s) no data", podDomain)
		}
		domainAllData.GeneratePlatformDataResult()
		c.setVTapPlatformData(domainAllData)
		log.Debug(domainAllData)
	} else if vtapConfig.PodClusterInternalIP == CLUSTER_OF_VTAP {
		// 下发的云平台列表=xxx，容器集群内部IP下发=采集器所在集群
		// 云平台列表=xxx中devicetype != POD/容器服务所有接口，集器所在集群devicetype=POD/容器服务的所有接口

		skipVifIDs := p.GetRawData().GetSkipVifIDs(c.GetLaunchServer())
		podDomains := c.getPodDomains()
		domainToPlatformDataExceptPod := p.GetDomainToPlatformDataExceptPod()
		domainToSkipPlatformDataExceptPod := p.GetDomainToSkipPlatformDataExceptPod()
		domainAllData := metadata.NewPlatformData("skipPlatformDataType3", "", 0, SKIP_PLATFORM_DATA_TYPE_3)
		for _, domainLcuuid := range vtapConfig.ConvertedDomains {
			if domainData, ok := domainToSkipPlatformDataExceptPod[domainLcuuid]; ok {
				domainAllData.SkipMerge(domainData, skipVifIDs)
				continue
			}
			if domainData, ok := domainToPlatformDataExceptPod[domainLcuuid]; ok {
				domainAllData.Merge(domainData)
				continue
			}
			log.Errorf("domain(%s) no platform data", domainLcuuid)
		}

		for _, podDomain := range podDomains {
			if vtapDomainData, ok := domainToSkipPlatformDataExceptPod[podDomain]; ok {
				if Find[string](vtapConfig.ConvertedDomains, podDomain) {
					domainAllData.SkipMergeInterfaces(vtapDomainData, skipVifIDs)
				} else {
					domainAllData.SkipMerge(vtapDomainData, skipVifIDs)
				}
				continue
			}
			if vtapDomainData, ok := domainToPlatformDataExceptPod[podDomain]; ok {
				if Find[string](vtapConfig.ConvertedDomains, podDomain) {
					domainAllData.MergeInterfaces(vtapDomainData)
				} else {
					domainAllData.Merge(vtapDomainData)
				}
				continue
			}
			log.Errorf("domain(%s) no platform data", podDomain)
		}

		domainAllData.GeneratePlatformDataResult()
		c.setVTapPlatformData(domainAllData)
		log.Debug(domainAllData)
	}
}
