package db

import (
	"math/rand"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"server/controller/db/mysql"
)

func newDBVMSecurityGroup() *mysql.VMSecurityGroup {
	return &mysql.VMSecurityGroup{Base: mysql.Base{Lcuuid: uuid.New().String()}, Priority: rand.Intn(10)}
}

func (t *SuiteTest) TestAddVMSecurityGroupBatchSuccess() {
	operator := NewVMSecurityGroup()
	itemToAdd := newDBVMSecurityGroup()

	_, ok := operator.AddBatch([]*mysql.VMSecurityGroup{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *mysql.VMSecurityGroup
	result := t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.VMSecurityGroup{})
}

func (t *SuiteTest) TestUpdateVMSecurityGroupSuccess() {
	operator := NewVMSecurityGroup()
	addedItem := newDBVMSecurityGroup()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	updateInfo := map[string]interface{}{"priority": rand.Intn(10)}
	_, ok := operator.Update(addedItem.Lcuuid, updateInfo)
	assert.True(t.T(), ok)

	var updatedItem *mysql.VMSecurityGroup
	t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), updatedItem.Priority, updateInfo["priority"])

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.VMSecurityGroup{})
}

func (t *SuiteTest) TestDeleteVMSecurityGroupBatchSuccess() {
	operator := NewVMSecurityGroup()
	addedItem := newDBVMSecurityGroup()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *mysql.VMSecurityGroup
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}
