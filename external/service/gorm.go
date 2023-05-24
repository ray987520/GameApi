package es

import (
	"TestAPI/external/service/zaplog"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type GormDB struct {
}

var (
	sqlDB            *gorm.DB
	sqlConnectString string
)

// 取GormDB實例
func GetSqlDb() *GormDB {
	return &GormDB{}
}

// 初始化,建立sql db連線與實例
func init() {
	sqlConnectString = "sqlserver://sa:0okmNJI(@localhost:1688?database=MYDB"
	db, err := gorm.Open(sqlserver.Open(sqlConnectString), &gorm.Config{})
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:           []gorm.Dialector{sqlserver.Open(sqlConnectString)},
		Replicas:          []gorm.Dialector{sqlserver.Open("sqlserver://sa:0okmNJI(@localhost:1688?database=MYDB")},
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}).
		SetMaxOpenConns(10).
		SetMaxIdleConns(2).
		SetConnMaxIdleTime(5 * time.Second))
	if err != nil {
		zaplog.Errorw("open gorm connection error", "error", err)
		panic(err)
	}
	sqlDb, err := db.DB()
	if err != nil {
		zaplog.Errorw("get gorm client error", "error", err)
		panic(err)
	}
	data, _ := json.Marshal(sqlDb.Stats())
	//初始化後列印DB狀態
	fmt.Println(string(data))
	sqlDB = db
}

// sql raw執行select
func (gormDB *GormDB) Select(model interface{}, sqlString string, params ...interface{}) error {
	tx := sqlDB.Raw(sqlString, params...).Scan(model)
	return tx.Error
}

// sql raw執行update
func (gormDB *GormDB) Update(sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	return tx.Error
}

// sql raw執行delete
func (gormDB *GormDB) Delete(sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	return tx.Error
}

// sql raw執行insert
func (gormDB *GormDB) Create(sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	return tx.Error
}

// sql執行batchinsert,使用gorm ORM
func (gormDB *GormDB) BatchCreate(tableName string, datas interface{}, batchSize int) error {
	tx := sqlDB.Table(tableName).CreateInBatches(datas, batchSize)
	return tx.Error
}

// sql raw執行transaction
func (gormDB *GormDB) Transaction(sqlStrings []string, params ...[]interface{}) error {
	tx := sqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for i, sql := range sqlStrings {
		if params != nil {
			partwork := tx.Exec(sql, params[i]...)
			if partwork.Error != nil {
				tx.Rollback()
				return partwork.Error
			}
		} else {
			partwork := tx.Exec(sql)
			if partwork.Error != nil {
				tx.Rollback()
				return partwork.Error
			}
		}
	}
	if tx.Error != nil {
		tx.Rollback()
	}
	return tx.Commit().Error
}
