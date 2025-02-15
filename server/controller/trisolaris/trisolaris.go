package trisolaris

import (
	"github.com/op/go-logging"
	"gorm.io/gorm"

	"server/controller/trisolaris/config"
	"server/controller/trisolaris/metadata"
	"server/controller/trisolaris/node"
	grpcserver "server/controller/trisolaris/server/grpc"
	httpserver "server/controller/trisolaris/server/http"
	"server/controller/trisolaris/utils"
	"server/controller/trisolaris/vtap"
)

var log = logging.MustGetLogger("trisolaris")

type Trisolaris struct {
	config   *config.Config
	dbConn   *gorm.DB
	metaData *metadata.MetaData
	vTapInfo *vtap.VTapInfo
	nodeInfo *node.NodeInfo
}

var trisolaris *Trisolaris

func GetGVTapInfo() *vtap.VTapInfo {
	return trisolaris.vTapInfo
}

func GetGNodeInfo() *node.NodeInfo {
	return trisolaris.nodeInfo
}

func GetConfig() *config.Config {
	return trisolaris.config
}

func GetDB() *gorm.DB {
	return trisolaris.dbConn
}

func PutPlatformData() {
	trisolaris.metaData.PutChPlatformData()
}

func PutTapType() {
	log.Info("PutTapType")
	trisolaris.metaData.PutChTapType()
}

func PutNodeInfo() {
	trisolaris.nodeInfo.PutChNodeInfo()
}

func PutVTapCache() {
	trisolaris.vTapInfo.PutVTapCacheRefresh()
}

func (t *Trisolaris) Start() {
	t.metaData.InitData() // 需要先初始化
	ctx, _ := utils.NewWaitGroupCtx()
	go t.metaData.TimedRefreshMetaData()
	go t.vTapInfo.TimedRefreshVTapCache()
	go t.nodeInfo.TimedRefreshNodeCache()
	go grpcserver.Run(ctx, t.config)
	go httpserver.Run(ctx, t.config)
}

func NewTrisolaris(cfg *config.Config, db *gorm.DB) *Trisolaris {
	if trisolaris == nil {
		cfg.Convert()
		metaData := metadata.NewMetaData(db, cfg)
		trisolaris = &Trisolaris{
			config:   cfg,
			dbConn:   db,
			metaData: metaData,
			vTapInfo: vtap.NewVTapInfo(db, metaData, cfg),
			nodeInfo: node.NewNodeInfo(db, metaData, cfg),
		}
	} else {
		return trisolaris
	}

	return trisolaris
}
