package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"server/controller/db/mysql"
)

func newDBPodNamespace() *mysql.PodNamespace {
	return &mysql.PodNamespace{Base: mysql.Base{Lcuuid: uuid.New().String()}, Region: uuid.New().String()}
}

func (t *SuiteTest) TestAddPodNamespaceBatchSuccess() {
	operator := NewPodNamespace()
	itemToAdd := newDBPodNamespace()

	_, ok := operator.AddBatch([]*mysql.PodNamespace{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *mysql.PodNamespace
	t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), addedItem.Lcuuid, itemToAdd.Lcuuid)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.PodNamespace{})
}

func (t *SuiteTest) TestUpdatePodNamespaceSuccess() {
	operator := NewPodNamespace()
	addedItem := newDBPodNamespace()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	updateInfo := map[string]interface{}{"region": uuid.New().String()}
	_, ok := operator.Update(addedItem.Lcuuid, updateInfo)
	assert.True(t.T(), ok)

	var updatedItem *mysql.PodNamespace
	t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), updatedItem.Region, updateInfo["region"])

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.PodNamespace{})
}

func (t *SuiteTest) TestDeletePodNamespaceBatchSuccess() {
	operator := NewPodNamespace()
	addedItem := newDBPodNamespace()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *mysql.PodNamespace
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}
