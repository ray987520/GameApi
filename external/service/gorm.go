package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
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
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlInit, innererror.ErrorTypeNode, innererror.InitGromError, innererror.ErrorInfoNode, err)
		panic(err)
	}
	sqlDb, err := db.DB()
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlInit, innererror.ErrorTypeNode, innererror.InitGromError, innererror.ErrorInfoNode, err)
		panic(err)
	}
	data, _ := json.Marshal(sqlDb.Stats())
	//初始化後列印DB狀態
	fmt.Println(string(data))
	sqlDB = db
}

// sql raw執行select
func (gormDB *GormDB) Select(traceMap string, model interface{}, sqlString string, params ...interface{}) error {
	tx := sqlDB.Raw(sqlString, params...).Scan(model)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlSelect, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.Error
}

// sql raw執行update
func (gormDB *GormDB) Update(traceMap string, sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlUpdate, innererror.ErrorTypeNode, innererror.UpdateError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.Error
}

// sql raw執行delete
func (gormDB *GormDB) Delete(traceMap string, sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlDelete, innererror.ErrorTypeNode, innererror.DeleteError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.Error
}

// sql raw執行insert
func (gormDB *GormDB) Create(traceMap string, sqlString string, params ...interface{}) error {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlCreate, innererror.ErrorTypeNode, innererror.CreateError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.Error
}

// sql執行batchinsert,使用gorm ORM
func (gormDB *GormDB) BatchCreate(traceMap string, tableName string, datas interface{}, batchSize int) error {
	tx := sqlDB.Table(tableName).CreateInBatches(datas, batchSize)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlBatchCreate, innererror.ErrorTypeNode, innererror.BatchCreateError, innererror.ErrorInfoNode, tx.Error, "tableName", tableName, "datas", datas)
	}
	return tx.Error
}

// sql raw執行transaction
func (gormDB *GormDB) Transaction(traceMap string, sqlStrings []string, params ...[]interface{}) error {
	tx := sqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	//循序執行所有sql
	for i, sql := range sqlStrings {
		if params != nil {
			partwork := tx.Exec(sql, params[i]...)
			if partwork.Error != nil {
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.ErrorTypeNode, innererror.TransactionError, innererror.ErrorInfoNode, partwork.Error, "sql", sql, "params", params[i])
				tx.Rollback()
				return partwork.Error
			}
		} else {
			partwork := tx.Exec(sql)
			if partwork.Error != nil {
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.ErrorTypeNode, innererror.TransactionError, innererror.ErrorInfoNode, partwork.Error, "sql", sql)
				tx.Rollback()
				return partwork.Error
			}
		}
	}
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlCreate, innererror.ErrorTypeNode, innererror.CreateError, innererror.ErrorInfoNode, tx.Error)
		tx.Rollback()
	}
	return tx.Commit().Error
}
