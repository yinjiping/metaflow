package common

import (
	"database/sql"
	"fmt"
	"net"

	clickhouse "github.com/ClickHouse/clickhouse-go"
	logging "github.com/op/go-logging"
	"gitlab.yunshan.net/yunshan/droplet-libs/datatype"
	"gitlab.yunshan.net/yunshan/droplet-libs/grpc"
)

var log = logging.MustGetLogger("common")

func NewCKConnection(addr, username, password string) (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", fmt.Sprintf("%s?username=%s&password=%s", addr, username, password))
	if err != nil {
		return nil, err
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Warningf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return connect, nil
}

// 如果通过MAC匹配平台信息失败，则需要通过IP再获取, 解决工单122/126问题
func RegetInfoFromIP(isIPv6 bool, ip6 net.IP, ip4 uint32, epcID int16, platformData *grpc.PlatformInfoTable) *grpc.Info {
	if isIPv6 {
		return platformData.QueryIPV6Infos(epcID, ip6)
	} else {
		return platformData.QueryIPV4Infos(epcID, ip4)
	}
}

const (
	IpType         = 255
	InternatIpType = 0

	PodType     = 10
	PodNodeType = 14

	PodGroupType = 101
	ServiceType  = 102
)

func GetResourceGl0(podID, podNodeID, l3DeviceID uint32, l3DeviceType uint8, l3EpcID int16) (uint32, uint8) {
	if podID > 0 {
		return podID, PodType
	} else if podNodeID > 0 {
		return podNodeID, PodNodeType
	} else if l3DeviceID > 0 {
		return l3DeviceID, l3DeviceType
	} else if l3EpcID == datatype.EPC_FROM_INTERNET {
		return 0, InternatIpType
	}

	return 0, IpType
}

func GetResourceGl1(podGroupID, podNodeID, l3DeviceID uint32, l3DeviceType uint8, l3EpcID int16) (uint32, uint8) {
	if podGroupID > 0 {
		return podGroupID, PodGroupType
	} else if podNodeID > 0 {
		return podNodeID, PodNodeType
	} else if l3DeviceID > 0 {
		return l3DeviceID, l3DeviceType
	} else if l3EpcID == datatype.EPC_FROM_INTERNET {
		return 0, InternatIpType
	}
	return 0, IpType
}

func GetResourceGl2(serviceID, podGroupID, podNodeID, l3DeviceID uint32, l3DeviceType uint8, l3EpcID int16) (uint32, uint8) {
	if serviceID > 0 {
		return serviceID, ServiceType
	}
	return GetResourceGl1(podGroupID, podNodeID, l3DeviceID, l3DeviceType, l3EpcID)
}
