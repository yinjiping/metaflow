package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"server/controller/db/mysql"
)

func newDBVM() *mysql.VM {
	return &mysql.VM{Base: mysql.Base{Lcuuid: uuid.New().String()}, Name: uuid.New().String()}
}

func (t *SuiteTest) TestAddVMBatchSuccess() {
	operator := NewVM()
	itemToAdd := newDBVM()

	_, ok := operator.AddBatch([]*mysql.VM{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *mysql.VM
	t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), addedItem.Name, itemToAdd.Name)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.VM{})
}

func (t *SuiteTest) TestUpdateVMSuccess() {
	operator := NewVM()
	addedItem := newDBVM()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	updateInfo := map[string]interface{}{"name": uuid.New().String()}
	_, ok := operator.Update(addedItem.Lcuuid, updateInfo)
	assert.True(t.T(), ok)

	var updatedItem *mysql.VM
	t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), updatedItem.Name, updateInfo["name"])

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.VM{})
}

func (t *SuiteTest) TestDeleteVMBatchSuccess() {
	operator := NewVM()
	addedItem := newDBVM()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))
	vmPodNodeConn := newDBVMPodNodeConnection()
	vmPodNodeConn.VMID = addedItem.ID
	result = t.db.Create(&vmPodNodeConn)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *mysql.VM
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
	var deletedVMPodNodeConn *mysql.VMPodNodeConnection
	result = t.db.Where("lcuuid = ?", vmPodNodeConn.Lcuuid).Find(&deletedVMPodNodeConn)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}
