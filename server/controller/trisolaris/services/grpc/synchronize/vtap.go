package synchronize

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"gitlab.yunshan.net/yunshan/metaflow/message/common"
	api "gitlab.yunshan.net/yunshan/metaflow/message/trident"
	context "golang.org/x/net/context"

	. "server/controller/common"
	"server/controller/trisolaris"
	. "server/controller/trisolaris/common"
	"server/controller/trisolaris/pushmanager"
	. "server/controller/trisolaris/utils"
	"server/controller/trisolaris/vtap"
)

var (
	RAW_UDP = api.SocketType_RAW_UDP
	TCP     = api.SocketType_TCP
	UDP     = api.SocketType_UDP
)

var SOCKET_TYPE_TO_MESSAGE = map[string]api.SocketType{
	"RAW_UDP": RAW_UDP,
	"TCP":     TCP,
	"UDP":     UDP,
}

type VTapEvent struct{}

func NewVTapEvent() *VTapEvent {
	return &VTapEvent{}
}

func Int2Bool(i int) bool {
	if i == 0 {
		return false
	}

	return true
}

func (e *VTapEvent) generateConfigInfo(c *vtap.VTapCache) *api.Config {
	gVTapInfo := trisolaris.GetGVTapInfo()
	proxyControllerIP := c.GetControllerIP()
	vtapConfig := c.GetVTapConfig()
	if vtapConfig == nil {
		return &api.Config{}
	}
	if vtapConfig.NatIPEnabled == 1 {
		proxyControllerIP = trisolaris.GetGNodeInfo().GetControllerNatIP(proxyControllerIP)
	}
	collectorSocketType, ok := SOCKET_TYPE_TO_MESSAGE[vtapConfig.CollectorSocketType]
	if ok == false {
		collectorSocketType = UDP
	}
	compressorSocketType, ok := SOCKET_TYPE_TO_MESSAGE[vtapConfig.CompressorSocketType]
	if ok == false {
		compressorSocketType = RAW_UDP
	}
	npbSocketType, ok := SOCKET_TYPE_TO_MESSAGE[vtapConfig.NpbSocketType]
	if ok == false {
		npbSocketType = RAW_UDP
	}
	decapTypes := make([]api.DecapType, 0, len(vtapConfig.ConvertedDecapType))
	for _, decap := range vtapConfig.ConvertedDecapType {
		decapTypes = append(decapTypes, api.DecapType(decap))
	}
	npbVlanMode := api.VlanMode(vtapConfig.NpbVlanMode)
	ifMacSource := api.IfMacSource(vtapConfig.IfMacSource)
	captureSocketType := api.CaptureSocketType(vtapConfig.CaptureSocketType)
	vtapID := uint32(c.GetVTapID())
	tridentType := common.TridentType(c.GetVTapType())
	podClusterId := uint32(c.GetPodClusterID())
	vpcID := uint32(c.GetVPCID())
	configure := &api.Config{
		CollectorEnabled:              proto.Bool(Int2Bool(vtapConfig.CollectorEnabled)),
		CollectorSocketType:           &collectorSocketType,
		CompressorSocketType:          &compressorSocketType,
		PlatformEnabled:               proto.Bool(Int2Bool(vtapConfig.PlatformEnabled)),
		MaxCpus:                       proto.Uint32(uint32(vtapConfig.MaxCPUs)),
		MaxMemory:                     proto.Uint32(uint32(vtapConfig.MaxMemory)),
		StatsInterval:                 proto.Uint32(uint32(vtapConfig.StatsInterval)),
		SyncInterval:                  proto.Uint32(uint32(vtapConfig.SyncInterval)),
		NpbBpsThreshold:               proto.Uint64(uint64(vtapConfig.MaxNpbBps)),
		GlobalPpsThreshold:            proto.Uint64(uint64(vtapConfig.MaxCollectPps)),
		Mtu:                           proto.Uint32(uint32(vtapConfig.Mtu)),
		OutputVlan:                    proto.Uint32(uint32(vtapConfig.OutputVlan)),
		RsyslogEnabled:                proto.Bool(Int2Bool(vtapConfig.RsyslogEnabled)),
		ServerTxBandwidthThreshold:    proto.Uint64(uint64(vtapConfig.MaxTxBandwidth)),
		BandwidthProbeInterval:        proto.Uint64(uint64(vtapConfig.BandwidthProbeInterval)),
		MaxEscapeSeconds:              proto.Uint32(uint32(vtapConfig.MaxEscapeSeconds)),
		NpbVlanMode:                   &npbVlanMode,
		NpbDedupEnabled:               proto.Bool(Int2Bool(vtapConfig.NpbDedupEnabled)),
		IfMacSource:                   &ifMacSource,
		NpbSocketType:                 &npbSocketType,
		VtapFlow_1SEnabled:            proto.Bool(Int2Bool(vtapConfig.VTapFlow1sEnabled)),
		CapturePacketSize:             proto.Uint32(uint32(vtapConfig.CapturePacketSize)),
		InactiveServerPortEnabled:     proto.Bool(Int2Bool(vtapConfig.InactiveServerPortEnabled)),
		LibvirtXmlPath:                proto.String(vtapConfig.VMXMLPath),
		LogThreshold:                  proto.Uint32(uint32(vtapConfig.LogThreshold)),
		LogLevel:                      proto.String(vtapConfig.LogLevel),
		LogRetention:                  proto.Uint32(uint32(vtapConfig.LogRetention)),
		L4LogCollectNpsThreshold:      proto.Uint64(uint64(vtapConfig.L4LogCollectNpsThreshold)),
		L7LogCollectNpsThreshold:      proto.Uint64(uint64(vtapConfig.L7LogCollectNpsThreshold)),
		L7MetricsEnabled:              proto.Bool(Int2Bool(vtapConfig.L7MetricsEnabled)),
		L7LogPacketSize:               proto.Uint32(uint32(vtapConfig.L7LogPacketSize)),
		DecapType:                     decapTypes,
		CaptureSocketType:             &captureSocketType,
		CaptureBpf:                    proto.String(vtapConfig.CaptureBpf),
		ThreadThreshold:               proto.Uint32(uint32(vtapConfig.ThreadThreshold)),
		ProcessThreshold:              proto.Uint32(uint32(vtapConfig.ProcessThreshold)),
		HttpLogProxyClient:            proto.String(vtapConfig.HTTPLogProxyClient),
		HttpLogTraceId:                proto.String(vtapConfig.HTTPLogTraceID),
		HttpLogSpanId:                 proto.String(vtapConfig.HTTPLogSpanID),
		HttpLogXRequestId:             proto.String(vtapConfig.HTTPLogXRequestID),
		NtpEnabled:                    proto.Bool(Int2Bool(vtapConfig.NtpEnabled)),
		L4PerformanceEnabled:          proto.Bool(Int2Bool(vtapConfig.L4PerformanceEnabled)),
		KubernetesApiEnabled:          proto.Bool(false),
		SysFreeMemoryLimit:            proto.Uint32(uint32(vtapConfig.SysFreeMemoryLimit)),
		LogFileSize:                   proto.Uint32(uint32(vtapConfig.LogFileSize)),
		ExternalAgentHttpProxyEnabled: proto.Bool(Int2Bool(vtapConfig.ExternalAgentHTTPProxyEnabled)),
		ExternalAgentHttpProxyPort:    proto.Uint32(uint32(vtapConfig.ExternalAgentHTTPProxyPort)),
		// 调整后采集器配置信息
		L7LogStoreTapTypes: vtapConfig.ConvertedL7LogStoreTapTypes,
		L4LogTapTypes:      vtapConfig.ConvertedL4LogTapTypes,
		// 采集器其他配置
		Enabled:           proto.Bool(Int2Bool(c.GetVTapEnabled())),
		Host:              proto.String(c.GetVTapHost()),
		ProxyControllerIp: &proxyControllerIP,
		VtapId:            &vtapID,
		TridentType:       &tridentType,
		EpcId:             &vpcID,
		// 容器采集器所在容器集群ID
		PodClusterId: &podClusterId,
	}

	cacheTSBIP := c.GetTSDBIP()
	configTSDBIP := gVTapInfo.GetConfigTSDBIP()
	if configTSDBIP != "" {
		configure.AnalyzerIp = &configTSDBIP
	} else if cacheTSBIP != "" {
		if vtapConfig.NatIPEnabled == 0 {
			configure.AnalyzerIp = &cacheTSBIP
		} else {
			natIP := trisolaris.GetGNodeInfo().GetTSDBNatIP(cacheTSBIP)
			configure.AnalyzerIp = &natIP
		}
	}

	if configure.GetProxyControllerIp() == "" {
		log.Errorf("vtap(%s) has no proxy_controller_ip", c.GetCtrlIP())
	}
	if configure.GetAnalyzerIp() == "" {
		log.Errorf("vtap(%s) has no analyzer_ip", c.GetCtrlIP())
	}
	regionID := trisolaris.GetGNodeInfo().GetRegionIDByTSDBIP(c.GetTSDBIP())
	if regionID != 0 {
		configure.RegionId = &regionID
	}

	if vtapConfig.TapInterfaceRegex != "" {
		configure.TapInterfaceRegex = proto.String(vtapConfig.TapInterfaceRegex)
	}
	pcapDataRetention := trisolaris.GetGNodeInfo().GetPcapDataRetention()
	if pcapDataRetention != 0 {
		configure.PcapDataRetention = proto.Uint32(pcapDataRetention)
	}

	return configure
}

func (e *VTapEvent) generateConfigFile(c *vtap.VTapCache, in *api.SyncRequest) *api.LocalConfigFile {
	if Find[int](vtap.LOCAL_FILE_VTAP, int(c.GetVTapType())) {
		return &api.LocalConfigFile{
			LocalConfig: proto.String(""),
			Revision:    proto.String(c.GetConfigFileRevison()),
		}
	}
	configFile := c.GetConfigFile()
	configFileRevision := c.GetConfigFileRevison()
	localConfigFile := in.GetLocalConfigFile()
	rpcRevision := ""
	if localConfigFile != nil {
		rpcRevision = localConfigFile.GetRevision()
	}
	if configFile != "" && configFileRevision != rpcRevision {
		return &api.LocalConfigFile{
			LocalConfig: proto.String(configFile),
			Revision:    proto.String(configFileRevision),
		}
	}
	return &api.LocalConfigFile{
		LocalConfig: proto.String(""),
		Revision:    proto.String(c.GetConfigFileRevison()),
	}
}

func (e *VTapEvent) Sync(ctx context.Context, in *api.SyncRequest) (*api.SyncResponse, error) {
	gVTapInfo := trisolaris.GetGVTapInfo()
	ctrlIP := in.GetCtrlIp()
	ctrlMac := in.GetCtrlMac()
	vtapCacheKey := ctrlIP + "-" + ctrlMac
	vtapCache, err := e.getVTapCache(in)
	if err != nil {
		log.Warningf("err:%s ctrlIp is %s, ctrlMac is %s, hostIps is %s, name:%s,  revision:%s,  bootTime:%d",
			err, ctrlIP, ctrlMac, in.GetHostIps(), in.GetProcessName(), in.GetRevision(), in.GetBootTime())
		return &api.SyncResponse{
			Status:        &STATUS_FAILED,
			Revision:      proto.String(in.GetRevision()),
			SelfUpdateUrl: proto.String(gVTapInfo.GetSelfUpdateUrl()),
		}, nil
	}
	if vtapCache == nil {
		log.Warningf("vtap (ctrl_ip: %s, ctrl_mac: %s, host_ips: %s, kubernetes_cluster_id: %s) not found in cache. "+
			"NAME:%s  REVISION:%s  BOOT_TIME:%d",
			ctrlIP, ctrlMac, in.GetHostIps(), in.GetKubernetesClusterId(),
			in.GetProcessName(), in.GetRevision(), in.GetBootTime())

		gVTapInfo.Register(
			int(in.GetTapMode()),
			in.GetCtrlIp(),
			in.GetCtrlMac(),
			in.GetHostIps(),
			in.GetHost(),
			in.GetVtapGroupIdRequest())
		return e.noVTapResponse(in), nil
	}

	versionPlatformData := vtapCache.GetSimplePlatformDataVersion()
	if versionPlatformData != in.GetVersionPlatformData() {
		log.Infof("ctrl_ip is %s, ctrl_mac is %s, host_ips is %s, "+
			"(platform data version  %d -> %d), NAME:%s  REVISION:%s  BOOT_TIME:%d",
			ctrlIP, ctrlMac, in.GetHostIps(), versionPlatformData,
			in.GetVersionPlatformData(), in.GetProcessName(), in.GetRevision(),
			in.GetBootTime())
	} else {
		log.Debugf("ctrl_ip is %s, ctrl_mac is %s, host_ips is %s,"+
			"(platform data version  %d -> %d), NAME:%s  REVISION:%s  BOOT_TIME:%d",
			ctrlIP, ctrlMac, in.GetHostIps(), versionPlatformData, in.GetVersionPlatformData(),
			in.GetProcessName(), in.GetRevision(), in.GetBootTime())
	}

	// trident上报的revision与yaml文件中trident_revision一致后，则取消预期的`config_revision`
	if vtapCache.GetUpgradeRevision() == in.GetRevision() {
		vtapCache.UpdateUpgradeRevision("")
	}
	if uint32(vtapCache.GetBootTime()) != in.GetBootTime() {
		vtapCache.UpdateBootTime(in.GetBootTime())
	}
	if vtapCache.GetRevision() != in.GetRevision() {
		vtapCache.UpdateRevision(in.GetRevision())
	}
	tridentException := vtapCache.GetExceptions() & VTAP_TRIDENT_EXCEPTIONS_MASK
	if tridentException != int64(in.GetException()) {
		vtapCache.UpdateExceptions(int64(in.GetException()))
	}
	vtapCache.UpdateSyncedControllerAt(time.Now())
	vtapCache.UpdateSystemInfoFromGrpc(
		int(in.GetCpuNum()),
		int64(in.GetMemorySize()),
		in.GetArch(),
		in.GetOs(),
		in.GetKernelVersion())
	localConfigFile := in.GetLocalConfigFile()
	if localConfigFile != nil {
		vtapCache.UpdateConfigFileFromGrpc(localConfigFile.GetLocalConfig(), localConfigFile.GetRevision())
	}
	// 专属采集器ctrl_mac可能会变，不更新ctrl_mac
	if vtapCache.GetVTapType() != VTAP_TYPE_DEDICATED {
		vtapCache.UpdateCtrlMacFromGrpc(in.GetCtrlMac())
	}
	vtapCache.SetControllerSyncFlag()
	// 记录采集器版本号， push接口用
	if in.GetVersionPlatformData() != 0 {
		vtapCache.UpdatePushVersionPlatformData(in.GetVersionPlatformData())
	} else {
		vtapCache.UpdatePushVersionPlatformData(versionPlatformData)
	}
	platformData := []byte{}
	if versionPlatformData != in.GetVersionPlatformData() {
		platformData = vtapCache.GetSimplePlatformDataStr()
	}

	// 只有专属采集器下发tap_types
	tapTypes := []*api.TapType{}
	if vtapCache.GetVTapType() == VTAP_TYPE_DEDICATED {
		tapTypes = gVTapInfo.GetTapTypes()
	}

	configInfo := e.generateConfigInfo(vtapCache)
	// 携带信息有cluster_id时选择一个采集器开启云平台同步开关
	if in.GetKubernetesClusterId() != "" {
		value := gVTapInfo.GetKubernetesClusterID(in.GetKubernetesClusterId(), vtapCacheKey)
		if value == vtapCacheKey {
			log.Infof(
				"open cluster(%s) kubernetes_api_enabled VTap(ctrl_ip: %s, ctrl_mac: %s)",
				in.GetKubernetesClusterId(), ctrlIP, ctrlMac)
			configInfo.KubernetesApiEnabled = proto.Bool(true)
		}
	}
	configFile := e.generateConfigFile(vtapCache, in)
	if in.GetLocalConfigFile() != nil && in.GetLocalConfigFile().GetRevision() != configFile.GetRevision() {
		log.Infof("send vtap(ctrl_ip: %s, ctrl_mac: %s), file_revision: %s file_len: %d",
			ctrlIP, ctrlMac, configFile.GetRevision(), len(configFile.GetLocalConfig()))
	} else {
		log.Debugf("send vatp(ctrl_ip: %s, ctrl_mac: %s), file_revision: %s file_len: %d",
			ctrlIP, ctrlMac, configFile.GetRevision(), len(configFile.GetLocalConfig()))
	}
	localSegments := vtapCache.GetVTapLocalSegments()
	remoteSegments := vtapCache.GetVTapRemoteSegments()
	upgradeRevision := vtapCache.GetUpgradeRevision()
	return &api.SyncResponse{
		Status:              &STATUS_SUCCESS,
		LocalSegments:       localSegments,
		RemoteSegments:      remoteSegments,
		Config:              configInfo,
		PlatformData:        platformData,
		VersionPlatformData: proto.Uint64(versionPlatformData),
		TapTypes:            tapTypes,
		LocalConfigFile:     configFile,
		SelfUpdateUrl:       proto.String(gVTapInfo.GetSelfUpdateUrl()),
		Revision:            proto.String(upgradeRevision),
	}, nil
}

func (e *VTapEvent) noVTapResponse(in *api.SyncRequest) *api.SyncResponse {
	ctrlIP := in.GetCtrlIp()
	ctrlMac := in.GetCtrlMac()
	vtapCacheKey := ctrlIP + "-" + ctrlMac
	gVTapInfo := trisolaris.GetGVTapInfo()
	if in.GetKubernetesClusterId() != "" {
		value := gVTapInfo.GetKubernetesClusterID(in.GetKubernetesClusterId(), vtapCacheKey)
		if value == vtapCacheKey {
			tridentType := common.TridentType(VTAP_TYPE_POD_VM)
			configInfo := &api.Config{
				KubernetesApiEnabled: proto.Bool(true),
				AnalyzerIp:           proto.String("127.0.0.1"),
				MaxEscapeSeconds:     proto.Uint32(uint32(gVTapInfo.GetDefaultMaxEscapeSeconds())),
				MaxMemory:            proto.Uint32(uint32(gVTapInfo.GetDefaultMaxMemory())),
				Enabled:              proto.Bool(true),
				TridentType:          &tridentType,
			}
			log.Infof(
				"open cluster(%s) kubernetes_api_enabled VTap(ctrl_ip: %s, ctrl_mac: %s)",
				in.GetKubernetesClusterId(), ctrlIP, ctrlMac)
			return &api.SyncResponse{
				Status: &STATUS_SUCCESS,
				Config: configInfo,
			}
		}
	}

	tridentTypeForUnkonwVTap := gVTapInfo.GetTridentTypeForUnkonwVTap()
	if tridentTypeForUnkonwVTap != 0 || in.GetTapMode() == api.TapMode_LOCAL {
		tridentType := common.TridentType(VTAP_TYPE_KVM)
		if tridentTypeForUnkonwVTap != 0 {
			tridentType = common.TridentType(tridentTypeForUnkonwVTap)
		}
		configInfo := &api.Config{
			TridentType:      &tridentType,
			AnalyzerIp:       proto.String("127.0.0.1"),
			MaxEscapeSeconds: proto.Uint32(uint32(gVTapInfo.GetDefaultMaxEscapeSeconds())),
			MaxMemory:        proto.Uint32(uint32(gVTapInfo.GetDefaultMaxMemory())),
			PlatformEnabled:  proto.Bool(true),
		}

		return &api.SyncResponse{
			Status: &STATUS_SUCCESS,
			Config: configInfo,
		}
	}
	return &api.SyncResponse{
		Status: &STATUS_FAILED,
	}
}

func (e *VTapEvent) getVTapCache(in *api.SyncRequest) (*vtap.VTapCache, error) {
	gVTapInfo := trisolaris.GetGVTapInfo()
	ctrlIP := in.GetCtrlIp()
	ctrlMac := in.GetCtrlMac()
	vtapCacheKey := ctrlIP + "-" + ctrlMac
	if !gVTapInfo.GetVTapCacheIsReady() {
		return nil, fmt.Errorf("VTap cache data not ready")
	}

	vtapCache := gVTapInfo.GetVTapCache(vtapCacheKey)
	if vtapCache == nil {
		vtapCache = gVTapInfo.GetVTapCache(ctrlIP)
		if vtapCache == nil {
			vtapCache = gVTapInfo.GetKvmVTapCache(ctrlIP)
			// ctrl_ip是kvm采集器的，但是ctrl_mac不属于tap_ports，需自动发现采集器
			if vtapCache != nil && gVTapInfo.IsCtrlMacInTapPorts(ctrlIP, ctrlMac) == false {
				vtapCache = nil
			}
		}
	}
	return vtapCache, nil
}

func (e *VTapEvent) pushResponse(in *api.SyncRequest) (*api.SyncResponse, error) {
	ctrlIP := in.GetCtrlIp()
	ctrlMac := in.GetCtrlMac()
	vtapCacheKey := ctrlIP + "-" + ctrlMac
	gVTapInfo := trisolaris.GetGVTapInfo()
	vtapCache, err := e.getVTapCache(in)
	if err != nil {
		return &api.SyncResponse{
			Status:        &STATUS_FAILED,
			Revision:      proto.String(in.GetRevision()),
			SelfUpdateUrl: proto.String(gVTapInfo.GetSelfUpdateUrl()),
		}, err
	}
	if vtapCache == nil {
		return e.noVTapResponse(in), fmt.Errorf("no find vtap(%s %s) cache", ctrlIP, ctrlMac)
	}

	versionPlatformData := vtapCache.GetPushVersionPlatformData()
	if versionPlatformData != in.GetVersionPlatformData() {
		log.Infof("push data ctrl_ip is %s, ctrl_mac is %s, host_ips is %s, "+
			"(platform data version  %d -> %d), NAME:%s  REVISION:%s  BOOT_TIME:%d",
			ctrlIP, ctrlMac, in.GetHostIps(), versionPlatformData,
			in.GetVersionPlatformData(), in.GetProcessName(), in.GetRevision(),
			in.GetBootTime())
	} else {
		log.Debugf("push data ctrl_ip is %s, ctrl_mac is %s, host_ips is %s,"+
			"(platform data version  %d -> %d), NAME:%s  REVISION:%s  BOOT_TIME:%d",
			ctrlIP, ctrlMac, in.GetHostIps(), versionPlatformData, in.GetVersionPlatformData(),
			in.GetProcessName(), in.GetRevision(), in.GetBootTime())
	}

	platformData := []byte{}
	if versionPlatformData != in.GetVersionPlatformData() {
		platformData = vtapCache.GetSimplePlatformDataStr()
	}

	// 只有专属采集器下发tap_types
	tapTypes := []*api.TapType{}
	if vtapCache.GetVTapType() == VTAP_TYPE_DEDICATED {
		tapTypes = gVTapInfo.GetTapTypes()
	}

	configInfo := e.generateConfigInfo(vtapCache)
	// 携带信息有cluster_id时选择一个采集器开启云平台同步开关
	if in.GetKubernetesClusterId() != "" {
		value := gVTapInfo.GetKubernetesClusterID(in.GetKubernetesClusterId(), vtapCacheKey)
		if value == vtapCacheKey {
			log.Infof(
				"open cluster(%s) kubernetes_api_enabled VTap(ctrl_ip: %s, ctrl_mac: %s)",
				in.GetKubernetesClusterId(), ctrlIP, ctrlMac)
			configInfo.KubernetesApiEnabled = proto.Bool(true)
		}
	}
	configFile := e.generateConfigFile(vtapCache, in)
	localSegments := vtapCache.GetVTapLocalSegments()
	remoteSegments := vtapCache.GetVTapRemoteSegments()
	return &api.SyncResponse{
		Status:              &STATUS_SUCCESS,
		LocalSegments:       localSegments,
		RemoteSegments:      remoteSegments,
		Config:              configInfo,
		PlatformData:        platformData,
		VersionPlatformData: proto.Uint64(versionPlatformData),
		TapTypes:            tapTypes,
		LocalConfigFile:     configFile,
	}, nil
}

func (e *VTapEvent) Push(r *api.SyncRequest, in api.Synchronizer_PushServer) error {
	var err error
	for {
		response, err := e.pushResponse(r)
		if err != nil {
			log.Error(err)
		}
		err = in.Send(response)
		if err != nil {
			log.Error(err)
			break
		}
		pushmanager.Wait()
	}
	log.Info("exit push", r.GetCtrlIp(), r.GetCtrlMac())
	return err
}
