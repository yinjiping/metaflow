package db

import (
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"server/controller/db/mysql"
)

func randomIP() string {
	return "192.168." + strconv.Itoa(rand.Intn(256)) + "." + strconv.Itoa(rand.Intn(256))
}

func newDBLANIP() *mysql.LANIP {
	return &mysql.LANIP{Base: mysql.Base{Lcuuid: uuid.New().String()}, IP: randomIP()}
}

func (t *SuiteTest) TestAddLANIPBatchSuccess() {
	operator := NewLANIP()
	itemToAdd := newDBLANIP()

	_, ok := operator.AddBatch([]*mysql.LANIP{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *mysql.LANIP
	t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), addedItem.IP, itemToAdd.IP)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.LANIP{})
}

func (t *SuiteTest) TestDeleteLANIPBatchSuccess() {
	operator := NewLANIP()
	addedItem := newDBLANIP()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *mysql.LANIP
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}
