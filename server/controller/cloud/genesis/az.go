package genesis

import (
	"server/controller/cloud/model"
	"server/controller/common"

	uuid "github.com/satori/go.uuid"
)

func (g *Genesis) getAZ() (model.AZ, error) {
	log.Debug("get az starting")
	azLcuuid := common.GetUUID(common.DEFAULT_REGION_NAME, uuid.Nil)
	az := model.AZ{
		Lcuuid:       azLcuuid,
		RegionLcuuid: g.regionUuid,
		Name:         common.DEFAULT_REGION_NAME,
	}
	g.azLcuuid = azLcuuid
	log.Debug("get az complete")
	return az, nil
}
