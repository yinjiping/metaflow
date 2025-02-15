package qingcloud

import (
	"sort"
	"strings"

	"github.com/deckarep/golang-set"

	"server/controller/cloud/model"
	"server/controller/common"
)

func (q *QingCloud) getRegionAndAZs() ([]model.Region, []model.AZ, error) {
	var retRegions []model.Region
	var retAZs []model.AZ
	var regionIdToLcuuid map[string]string
	var zoneNames []string

	log.Debug("get region and azs starting")

	kwargs := []*Param{{"status.1", "active"}}
	response, err := q.GetResponse("DescribeZones", "zone_set", kwargs)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	regionIds := mapset.NewSet()
	regionIdToLcuuid = make(map[string]string)
	for _, r := range response {
		for i := range r.MustArray() {
			zone := r.GetIndex(i)
			err := q.CheckRequiredAttributes(zone, []string{"zone_id"})
			if err != nil {
				continue
			}

			zoneId := zone.Get("zone_id").MustString()
			// 亚太2区-A和雅加达区都是ap开头，但是不能合并为一个区域
			regionId := zoneId[:len(zoneId)-1]
			if strings.HasPrefix(regionId, "ap") {
				regionId = zoneId
			}
			regionLcuuid := common.GenerateUUID(q.UuidGenerate + "_" + regionId)
			retAZs = append(retAZs, model.AZ{
				Lcuuid:       common.GenerateUUID(q.UuidGenerate + "_" + zoneId),
				Name:         zoneId,
				Label:        zoneId,
				RegionLcuuid: q.GetRegionLcuuid(regionLcuuid),
			})
			zoneNames = append(zoneNames, zoneId)

			// 生成区域列表
			if q.RegionUuid == "" {
				regionIds.Add(regionId)
				regionIdToLcuuid[regionId] = regionLcuuid
			} else {
				regionIdToLcuuid[regionId] = q.RegionUuid
			}

		}
	}
	sort.Strings(zoneNames)
	q.RegionIdToLcuuid = regionIdToLcuuid
	q.ZoneNames = zoneNames

	// 生成区域返回数据
	for _, regionId := range regionIds.ToSlice() {
		regionIdStr := regionId.(string)
		retRegions = append(retRegions, model.Region{
			Lcuuid: common.GenerateUUID(q.UuidGenerate + "_" + regionIdStr),
			Name:   strings.ToUpper(regionIdStr),
		})
	}

	log.Debug("get region and azs complete")
	return retRegions, retAZs, nil
}
