package genesis

import (
	"server/controller/cloud/model"
	"server/controller/common"
	"server/controller/genesis"
	"strconv"

	uuid "github.com/satori/go.uuid"
)

func (g *Genesis) getNetworks() ([]model.Network, error) {
	log.Debug("get networks starting")
	networks := []model.Network{}
	networksData := genesis.GenesisService.GetNetworksData()
	for _, n := range networksData {
		if n.SegmentationID == 0 {
			log.Debug("segmentation id not found")
			continue
		}
		vpcLcuuid := n.VPCLcuuid
		if vpcLcuuid == "" {
			vpcLcuuid = common.GetUUID(g.defaultVpcName, uuid.Nil)
			g.defaultVpc = true
		}
		networkName := n.Name
		if networkName == "" {
			networkName = "subnet_vni_" + strconv.Itoa(n.SegmentationID)
		}
		network := model.Network{
			Lcuuid:         n.Lcuuid,
			Name:           networkName,
			SegmentationID: n.SegmentationID,
			VPCLcuuid:      vpcLcuuid,
			Shared:         false,
			External:       n.External,
			NetType:        n.NetType,
			RegionLcuuid:   g.regionUuid,
		}
		networks = append(networks, network)
	}
	log.Debug("get networks complete")
	return networks, nil
}
