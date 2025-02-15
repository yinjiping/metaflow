package tagrecorder

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"server/controller/db/mysql"
)

const (
	TEST_DB_FILE = "./tagrecorder_test.db"
)

type SuiteTest struct {
	suite.Suite
	db *gorm.DB
}

func TestSuite(t *testing.T) {
	if _, err := os.Stat(TEST_DB_FILE); err == nil {
		os.Remove(TEST_DB_FILE)
	}
	mysql.Db = GetDB(TEST_DB_FILE)
	suite.Run(t, new(SuiteTest))
}

func (t *SuiteTest) SetupSuite() {
	t.db = mysql.Db

	for _, val := range getModels() {
		t.db.AutoMigrate(val)
	}
}

func (t *SuiteTest) TearDownSuite() {
	sqlDB, _ := t.db.DB()
	sqlDB.Close()

	os.Remove(TEST_DB_FILE)
}

func GetDB(dbFile string) *gorm.DB {
	db, err := gorm.Open(
		sqlite.Open(dbFile),
		&gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}},
	)
	if err != nil {
		fmt.Printf("create sqlite database failed: %s\n", err.Error())
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

func getModels() []interface{} {
	return []interface{}{
		&mysql.Region{}, &mysql.AZ{}, &mysql.VPC{}, &mysql.VM{}, &mysql.VInterface{},
		&mysql.WANIP{}, &mysql.LANIP{}, &mysql.NATGateway{}, &mysql.NATRule{},
		&mysql.NATVMConnection{}, &mysql.LB{}, &mysql.LBListener{}, &mysql.LBTargetServer{},
		&mysql.LBVMConnection{}, &mysql.PodIngress{}, &mysql.PodService{}, mysql.PodGroup{},
		&mysql.PodGroupPort{}, &mysql.Pod{},
		&mysql.ChRegion{}, &mysql.ChAZ{}, &mysql.ChVPC{}, &mysql.ChIPRelation{},
	}
}
