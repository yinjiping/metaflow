package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"server/controller/db/mysql"
)

func newNATRule() *mysql.NATRule {
	return &mysql.NATRule{Base: mysql.Base{Lcuuid: uuid.New().String()}}
}

func (t *SuiteTest) TestAddNATRuleBatchSuccess() {
	operator := NewNATRule()
	itemToAdd := newNATRule()

	_, ok := operator.AddBatch([]*mysql.NATRule{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *mysql.NATRule
	result := t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&mysql.NATRule{})
}

func (t *SuiteTest) TestDeleteNATRuleBatchSuccess() {
	operator := NewNATRule()
	addedItem := newNATRule()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *mysql.NATRule
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}
