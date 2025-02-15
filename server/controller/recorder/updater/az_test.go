package updater

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	cloudmodel "server/controller/cloud/model"
	"server/controller/db/mysql"
	"server/controller/recorder/cache"
)

func newCloudAZ() cloudmodel.AZ {
	lcuuid := uuid.New().String()
	return cloudmodel.AZ{
		Lcuuid:       lcuuid,
		Name:         lcuuid[:8],
		Label:        lcuuid[:6],
		RegionLcuuid: uuid.New().String(),
	}
}

func (t *SuiteTest) getAZMock(mockDB bool) (*cache.Cache, cloudmodel.AZ) {
	cloudItem := newCloudAZ()
	domainLcuuid := uuid.New().String()

	wholeCache := cache.NewCache(domainLcuuid)
	if mockDB {
		dbItem := new(mysql.AZ)
		dbItem.Lcuuid = cloudItem.Lcuuid
		dbItem.Name = cloudItem.Name
		t.db.Create(dbItem)
		wholeCache.AZs[cloudItem.Lcuuid] = &cache.AZ{DiffBase: cache.DiffBase{Lcuuid: cloudItem.Lcuuid}}
	}

	wholeCache.SetSequence(wholeCache.GetSequence() + 1)

	return wholeCache, cloudItem
}

func (t *SuiteTest) TestHandleAddAZSucess() {
	cache, cloudItem := t.getAZMock(false)
	assert.Equal(t.T(), len(cache.AZs), 0)

	updater := NewAZ(cache, []cloudmodel.AZ{cloudItem})
	updater.HandleAddAndUpdate()

	var addedItem *mysql.AZ
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), len(cache.AZs), 1)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.AZ{})
}

func (t *SuiteTest) TestHandleUpdateAZSucess() {
	cache, cloudItem := t.getAZMock(true)
	cloudItem.Name = cloudItem.Name + "new"

	updater := NewAZ(cache, []cloudmodel.AZ{cloudItem})
	updater.HandleAddAndUpdate()

	var updatedItem *mysql.AZ
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	assert.Equal(t.T(), updatedItem.Name, cloudItem.Name)
	assert.Equal(t.T(), cache.AZs[cloudItem.Lcuuid].Name, cloudItem.Name)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.AZ{})
}

func (t *SuiteTest) TestHandleDeleteAZSucess() {
	cache, cloudItem := t.getAZMock(true)

	updater := NewAZ(cache, []cloudmodel.AZ{})
	updater.HandleDelete()

	var addedItem *mysql.AZ
	result := t.db.Where("lcuuid = ?", cloudItem.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.AZ{})
}
