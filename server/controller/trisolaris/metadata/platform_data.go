package metadata

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/golang/protobuf/proto"
	"gitlab.yunshan.net/yunshan/metaflow/message/trident"
)

var offsetInterval uint64 = 1000000
var offsetVersion uint64 = 1000000

type PlatformData struct {
	domain             string
	lcuuid             string
	platformDataStr    []byte
	platformDataHash   uint64
	platformDataProtos *trident.PlatformData
	interfaceProtos    []*trident.Interface
	peerConnProtos     []*trident.PeerConnection
	cidrProtos         []*trident.Cidr
	version            uint64
	mergeDomains       []string
	dataType           uint32
}

func NewPlatformData(domain string, lcuuid string, version uint64, dataType uint32) *PlatformData {
	return &PlatformData{
		domain:             domain,
		lcuuid:             lcuuid,
		platformDataStr:    []byte{},
		platformDataHash:   0,
		platformDataProtos: &trident.PlatformData{},
		interfaceProtos:    []*trident.Interface{},
		peerConnProtos:     []*trident.PeerConnection{},
		cidrProtos:         []*trident.Cidr{},
		version:            version,
		mergeDomains:       []string{},
		dataType:           dataType,
	}
}

func (f *PlatformData) setPlatformData(ifs []*trident.Interface, pcs []*trident.PeerConnection, cidrs []*trident.Cidr) {
	f.initPlatformData(ifs, pcs, cidrs)
	f.GeneratePlatformDataResult()
}

func (f *PlatformData) GetPlatformDataResult() ([]byte, uint64) {
	return f.platformDataStr, f.version
}

func (f *PlatformData) GetPlatformDataStr() []byte {
	return f.platformDataStr
}

func (f *PlatformData) GetPlatformDataVersion() uint64 {
	return f.version
}

func (f *PlatformData) setVersion(version uint64) {
	f.version = version
}

func (f *PlatformData) GetVersion() uint64 {
	return f.version
}

func (f *PlatformData) initVersion() {
	rand.Seed(time.Now().Unix())
	f.version = offsetVersion + uint64(time.Now().Unix()) + uint64(rand.Intn(10000))
	offsetVersion += offsetInterval
}

func (f *PlatformData) initPlatformData(ifs []*trident.Interface, pcs []*trident.PeerConnection, cidrs []*trident.Cidr) {
	f.interfaceProtos = ifs
	f.peerConnProtos = pcs
	f.cidrProtos = cidrs
}

func (f *PlatformData) GeneratePlatformDataResult() {
	f.platformDataProtos = &trident.PlatformData{
		Interfaces:      f.interfaceProtos,
		PeerConnections: f.peerConnProtos,
		Cidrs:           f.cidrProtos,
	}
	var err error
	f.platformDataStr, err = f.platformDataProtos.Marshal()
	if err != nil {
		log.Error(err)
		return
	}
	h64 := fnv.New64()
	h64.Write(f.platformDataStr)
	f.platformDataHash = h64.Sum64()
}

func (f *PlatformData) GenerateSkipPlatformDataResult(skipVifIDs mapset.Set) {
	if skipVifIDs.Cardinality() > 0 {
		skipInterfaceProtos := make([]*trident.Interface, 0, len(f.interfaceProtos))
		for _, interfaceProto := range f.interfaceProtos {
			tInterfaceProto := proto.Clone(interfaceProto).(*trident.Interface)
			if skipVifIDs.Contains(int(interfaceProto.GetId())) {
				tInterfaceProto.SkipTapInterface = proto.Bool(true)
				skipInterfaceProtos = append(skipInterfaceProtos, tInterfaceProto)
			} else {
				tInterfaceProto.SkipTapInterface = proto.Bool(false)
				skipInterfaceProtos = append(skipInterfaceProtos, tInterfaceProto)
			}
		}
		f.platformDataProtos = &trident.PlatformData{
			Interfaces:      skipInterfaceProtos,
			PeerConnections: f.peerConnProtos,
			Cidrs:           f.cidrProtos,
		}
	} else {
		f.platformDataProtos = &trident.PlatformData{
			Interfaces:      f.interfaceProtos,
			PeerConnections: f.peerConnProtos,
			Cidrs:           f.cidrProtos,
		}
	}
	var err error
	f.platformDataStr, err = f.platformDataProtos.Marshal()
	if err != nil {
		log.Error(err)
		return
	}
	h64 := fnv.New64()
	h64.Write(f.platformDataStr)
	f.platformDataHash = h64.Sum64()
}

func (f *PlatformData) Merge(other *PlatformData) {
	f.interfaceProtos = append(f.interfaceProtos, other.interfaceProtos...)
	f.peerConnProtos = append(f.peerConnProtos, other.peerConnProtos...)
	f.cidrProtos = append(f.cidrProtos, other.cidrProtos...)
	f.version += other.version
	if len(other.domain) != 0 {
		f.mergeDomains = append(f.mergeDomains, other.domain)
	}
}

func (f *PlatformData) MergeInterfaces(other *PlatformData) {
	f.interfaceProtos = append(f.interfaceProtos, other.interfaceProtos...)
	f.version += other.version
	if len(other.domain) != 0 {
		f.mergeDomains = append(f.mergeDomains, other.domain)
	}
}

func (f *PlatformData) equal(other *PlatformData) bool {
	if f.platformDataHash != other.platformDataHash {
		return false
	}

	return true
}

func (f *PlatformData) String() string {
	return fmt.Sprintf("name: %s, lcuuid: %s, data_type: %d, version: %d, platform_data_hash: %d, interfaces: %d, peer_connections: %d, cidrs: %d, merge_domains: %s",
		f.domain, f.lcuuid, f.dataType, f.version, f.platformDataHash, len(f.interfaceProtos), len(f.peerConnProtos), len(f.cidrProtos), f.mergeDomains)
}
