package updater

import (
	"reflect"

	"bou.ke/monkey"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
)

func newCloudDHCPPort() cloudmodel.DHCPPort {
	lcuuid := uuid.NewString()
	return cloudmodel.DHCPPort{
		Lcuuid: lcuuid,
		Name:   lcuuid[:8],
	}
}

func (t *SuiteTest) getDHCPPortMock(mockDB bool) (*cache.Cache, cloudmodel.DHCPPort) {
	cloudItem := newCloudDHCPPort()
	domainLcuuid := uuid.NewString()

	cache_ := cache.NewCache(domainLcuuid)
	if mockDB {
		t.db.Create(&mysql.DHCPPort{Name: cloudItem.Name, Base: mysql.Base{Lcuuid: cloudItem.Lcuuid}, Domain: domainLcuuid})
		cache_.DHCPPorts[cloudItem.Lcuuid] = &cache.DHCPPort{DiffBase: cache.DiffBase{Lcuuid: cloudItem.Lcuuid}, Name: cloudItem.Name}
	}

	cache_.SetSequence(cache_.GetSequence() + 1)

	return cache_, cloudItem
}

func (t *SuiteTest) TestHandleAddDHCPPortSucess() {
	cache_, cloudItem := t.getDHCPPortMock(false)
	vpcID := randID()
	monkey.PatchInstanceMethod(reflect.TypeOf(&cache_.ToolDataSet), "GetVPCIDByLcuuid", func(_ *cache.ToolDataSet, _ string) (int, bool) {
		return vpcID, true
	})
	assert.Equal(t.T(), len(cache_.DHCPPorts), 0)

	updater := NewDHCPPort(cache_, []cloudmodel.DHCPPort{cloudItem})
	updater.HandleAddAndUpdate()

	monkey.UnpatchInstanceMethod(reflect.TypeOf(&cache_.ToolDataSet), "GetVPCIDByLcuuid")

	var addedItem *mysql.DHCPPort
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), len(cache_.DHCPPorts), 1)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.DHCPPort{})
}

func (t *SuiteTest) TestHandleUpdateDHCPPortSucess() {
	cache, cloudItem := t.getDHCPPortMock(true)
	cloudItem.Name = cloudItem.Name + "new"

	updater := NewDHCPPort(cache, []cloudmodel.DHCPPort{cloudItem})
	updater.HandleAddAndUpdate()

	var addedItem *mysql.DHCPPort
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), addedItem.Name, cloudItem.Name)
	assert.Equal(t.T(), len(cache.DHCPPorts), 1)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.DHCPPort{})
}

func (t *SuiteTest) TestHandleDeleteDHCPPortSucess() {
	cache, cloudItem := t.getDHCPPortMock(true)

	updater := NewDHCPPort(cache, []cloudmodel.DHCPPort{cloudItem})
	updater.HandleDelete()

	var addedItem *mysql.DHCPPort
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
	assert.Equal(t.T(), len(cache.DHCPPorts), 0)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.DHCPPort{})
}
