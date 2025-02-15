package common

import (
	"server/controller/db/mysql"
)

var DefaultVTapGroupConfig = &mysql.VTapGroupConfiguration{
	MaxCollectPps:                 &MaxCollectPps,
	MaxNpbBps:                     &MaxNpbBps,
	MaxCPUs:                       &MaxCPUs,
	MaxMemory:                     &MaxMemory,
	SyncInterval:                  &SyncInterval,
	StatsInterval:                 &StatsInterval,
	RsyslogEnabled:                &RsyslogEnabled,
	MaxTxBandwidth:                &MaxTxBandwidth,
	BandwidthProbeInterval:        &BandwidthProbeInterval,
	TapInterfaceRegex:             &TapInterfaceRegex,
	MaxEscapeSeconds:              &MaxEscapeSeconds,
	Mtu:                           &Mtu,
	OutputVlan:                    &OutputVlan,
	CollectorSocketType:           &CollectorSocketType,
	CompressorSocketType:          &CompressorSocketType,
	NpbSocketType:                 &NpbSocketType,
	NpbVlanMode:                   &NpbVlanMode,
	CollectorEnabled:              &CollectorEnabled,
	VTapFlow1sEnabled:             &VTapFlow1sEnabled,
	L4LogTapTypes:                 &L4LogTapTypes,
	NpbDedupEnabled:               &NpbDedupEnabled,
	PlatformEnabled:               &PlatformEnabled,
	IfMacSource:                   &IfMacSource,
	VMXMLPath:                     &VMXMLPath,
	NatIPEnabled:                  &NatIPEnabled,
	CapturePacketSize:             &CapturePacketSize,
	InactiveServerPortEnabled:     &InactiveServerPortEnabled,
	LogThreshold:                  &LogThreshold,
	LogLevel:                      &LogLevel,
	LogRetention:                  &LogRetention,
	HTTPLogProxyClient:            &HTTPLogProxyClient,
	HTTPLogTraceID:                &HTTPLogTraceID,
	L7LogPacketSize:               &L7LogPacketSize,
	L4LogCollectNpsThreshold:      &L4LogCollectNpsThreshold,
	L7LogCollectNpsThreshold:      &L7LogCollectNpsThreshold,
	L7MetricsEnabled:              &L7MetricsEnabled,
	L7LogStoreTapTypes:            &L7LogStoreTapTypes,
	CaptureSocketType:             &CaptureSocketType,
	CaptureBpf:                    &CaptureBpf,
	ThreadThreshold:               &ThreadThreshold,
	ProcessThreshold:              &ProcessThreshold,
	NtpEnabled:                    &NtpEnabled,
	L4PerformanceEnabled:          &L4PerformanceEnabled,
	PodClusterInternalIP:          &PodClusterInternalIP,
	Domains:                       &Domains,
	DecapType:                     &DecapType,
	HTTPLogSpanID:                 &HTTPLogSpanID,
	SysFreeMemoryLimit:            &SysFreeMemoryLimit,
	LogFileSize:                   &LogFileSize,
	HTTPLogXRequestID:             &HTTPLogXRequestID,
	ExternalAgentHTTPProxyEnabled: &ExternalAgentHTTPProxyEnabled,
	ExternalAgentHTTPProxyPort:    &ExternalAgentHTTPProxyPort,
}

var (
	MaxCollectPps                 = 200000
	MaxNpbBps                     = int64(1000000000)
	MaxCPUs                       = 1
	MaxMemory                     = 768
	SyncInterval                  = 60
	StatsInterval                 = 60
	RsyslogEnabled                = 1
	MaxTxBandwidth                = int64(0)
	BandwidthProbeInterval        = 10
	TapInterfaceRegex             = "^tap.*"
	MaxEscapeSeconds              = 3600
	Mtu                           = 1500
	OutputVlan                    = 0
	CollectorSocketType           = "TCP"
	CompressorSocketType          = "TCP"
	NpbSocketType                 = "RAW_UDP"
	NpbVlanMode                   = 0
	CollectorEnabled              = 1
	VTapFlow1sEnabled             = 1
	L4LogTapTypes                 = "0"
	NpbDedupEnabled               = 1
	PlatformEnabled               = 0
	IfMacSource                   = 0
	VMXMLPath                     = "/etc/libvirt/qemu/"
	NatIPEnabled                  = 0
	CapturePacketSize             = 65535
	InactiveServerPortEnabled     = 1
	LogThreshold                  = 300
	LogLevel                      = "INFO"
	LogRetention                  = 30
	HTTPLogProxyClient            = "X-Forwarded-For"
	HTTPLogTraceID                = "关闭"
	L7LogPacketSize               = 256
	L4LogCollectNpsThreshold      = 10000
	L7LogCollectNpsThreshold      = 10000
	L7MetricsEnabled              = 1
	L7LogStoreTapTypes            = "0"
	CaptureSocketType             = 0
	CaptureBpf                    = ""
	ThreadThreshold               = 100
	ProcessThreshold              = 10
	NtpEnabled                    = 1
	L4PerformanceEnabled          = 1
	PodClusterInternalIP          = 0
	Domains                       = "0"
	DecapType                     = "0"
	HTTPLogSpanID                 = "关闭"
	SysFreeMemoryLimit            = 30
	LogFileSize                   = 1000
	HTTPLogXRequestID             = "关闭"
	ExternalAgentHTTPProxyEnabled = 0
	ExternalAgentHTTPProxyPort    = 8086
)
