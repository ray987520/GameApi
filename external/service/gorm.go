package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/zaplog"
	"encoding/json"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type GormDB struct {
}

var (
	sqlDB            *gorm.DB
	sqlConnectString = mconfig.GetString("sql.connectString.master")
	maxOpenConns     = mconfig.GetInt("sql.maxOpenConns")
	maxIdleConns     = mconfig.GetInt("sql.maxIdleConns")
	maxIdleSecond    = mconfig.GetDuration("sql.maxIdleSecond")
)

// 取GormDB實例
func GetSqlDb() *GormDB {
	return &GormDB{}
}

// 初始化,建立sql db連線與實例
func init() {
	db, err := gorm.Open(sqlserver.Open(sqlConnectString), &gorm.Config{})
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:           []gorm.Dialector{sqlserver.Open(sqlConnectString)},
		Replicas:          []gorm.Dialector{sqlserver.Open(sqlConnectString)},
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}).
		SetMaxOpenConns(maxOpenConns).
		SetMaxIdleConns(maxIdleConns).
		SetConnMaxIdleTime(maxIdleSecond * time.Second))
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
	zaplog.Info(string(data))

	sqlDB = db
}

// sql raw執行select,輸出生效行數
func (gormDB *GormDB) Select(traceMap string, model interface{}, sqlString string, params ...interface{}) (rowsAffect int64, err error) {
	tx := sqlDB.Raw(sqlString, params...).Scan(model)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlSelect, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.RowsAffected, tx.Error
}

// sql raw執行update,輸出生效行數
func (gormDB *GormDB) Update(traceMap string, sqlString string, params ...interface{}) (rowsAffect int64, err error) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlUpdate, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.UpdateError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.RowsAffected, tx.Error
}

// sql raw執行delete,輸出生效行數
func (gormDB *GormDB) Delete(traceMap string, sqlString string, params ...interface{}) (rowsAffect int64, err error) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlDelete, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.DeleteError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.RowsAffected, tx.Error
}

// sql raw執行insert,輸出生效行數
func (gormDB *GormDB) Create(traceMap string, sqlString string, params ...interface{}) (rowsAffect int64, err error) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlCreate, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.CreateError, innererror.ErrorInfoNode, tx.Error, "sqlString", sqlString, "params", params)
	}
	return tx.RowsAffected, tx.Error
}

// sql執行batchinsert,使用gorm ORM,輸出生效行數
func (gormDB *GormDB) BatchCreate(traceMap string, tableName string, datas interface{}, batchSize int) (rowsAffect int64, err error) {
	tx := sqlDB.Table(tableName).CreateInBatches(datas, batchSize)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlBatchCreate, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.BatchCreateError, innererror.ErrorInfoNode, tx.Error, "tableName", tableName, "datas", datas)
	}
	return tx.RowsAffected, tx.Error
}

/* GORM手動transaction
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
*/

// GORM自動transaction,sql raw執行,輸出生效行數
func (gormDB *GormDB) Transaction(traceMap string, sqlStrings []string, params ...[]interface{}) (rowsAffect int64, err error) {
	err = sqlDB.Transaction(func(tx *gorm.DB) error {
		//循序執行所有sql
		for i, sql := range sqlStrings {
			if params != nil {
				partwork := tx.Exec(sql, params[i]...)
				if partwork.Error != nil {
					zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.TransactionError, innererror.ErrorInfoNode, partwork.Error, "sql", sql, "params", params[i])
					return partwork.Error
				}
				rowsAffect += partwork.RowsAffected
			} else {
				partwork := tx.Exec(sql)
				if partwork.Error != nil {
					zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.TraceNode, traceMap, innererror.ErrorTypeNode, innererror.TransactionError, innererror.ErrorInfoNode, partwork.Error, "sql", sql)
					return partwork.Error
				}
				rowsAffect += partwork.RowsAffected
			}
		}

		return nil
	})
	return
}
