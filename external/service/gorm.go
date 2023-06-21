package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"fmt"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type GormDB struct {
}

const (
	connectionError = "gorm open connection error:%v"
	dbInstanceError = "gorm get sql instance error:%v"
	dbStat          = "gorm db stat:%v"
)

var (
	sqlDB            *gorm.DB
	sqlConnectString string
	maxOpenConns     int
	maxIdleConns     int
	maxIdleSecond    time.Duration
)

// 取GormDB實例
func GetSqlDb() *GormDB {
	SqlInit()
	return &GormDB{}
}

// 初始化,建立sql db連線與實例
func SqlInit() {
	sqlConnectString = mconfig.GetString("sql.connectString.master")
	maxOpenConns = mconfig.GetInt("sql.maxOpenConns")
	maxIdleConns = mconfig.GetInt("sql.maxIdleConns")
	maxIdleSecond = mconfig.GetDuration("sql.maxIdleSecond")
	//gorm連接sql server
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
		err = fmt.Errorf(connectionError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, err)
		panic(err)
	}

	//嘗試取出DB的connection pool,錯誤就是連線但連不上DB
	sqlDb, err := db.DB()
	if err != nil {
		err = fmt.Errorf(dbInstanceError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, err)
		panic(err)
	}

	//初始化後列印DB狀態
	data := JsonMarshal(tracer.DefaultTraceId, sqlDb.Stats())
	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, esid.SqlInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, fmt.Sprintf(dbStat, string(data)))

	//把connection傳給全域變數
	sqlDB = db
}

// sql raw執行select,輸出生效行數
func (gormDB *GormDB) Select(traceId string, model interface{}, sqlString string, params ...interface{}) (rowsAffect int64) {
	tx := sqlDB.Raw(sqlString, params...).Scan(model)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlSelect, innererror.TraceNode, traceId, innererror.DataNode, tx.Error)
		return -1
	}
	return tx.RowsAffected
}

// sql raw執行update,輸出生效行數
func (gormDB *GormDB) Update(traceId string, sqlString string, params ...interface{}) (rowsAffect int64) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlUpdate, innererror.TraceNode, traceId, innererror.DataNode, tx.Error)
		return -1
	}
	return tx.RowsAffected
}

// sql raw執行delete,輸出生效行數
func (gormDB *GormDB) Delete(traceId string, sqlString string, params ...interface{}) (rowsAffect int64) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlDelete, innererror.TraceNode, traceId, innererror.DataNode, tx.Error)
		return -1
	}
	return tx.RowsAffected
}

// sql raw執行insert,輸出生效行數
func (gormDB *GormDB) Create(traceId string, sqlString string, params ...interface{}) (rowsAffect int64) {
	tx := sqlDB.Exec(sqlString, params...)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlCreate, innererror.TraceNode, traceId, innererror.DataNode, tx.Error)
		return -1
	}
	return tx.RowsAffected
}

// sql執行batchinsert,使用gorm ORM,輸出生效行數
func (gormDB *GormDB) BatchCreate(traceId string, tableName string, datas interface{}, batchSize int) (rowsAffect int64) {
	tx := sqlDB.Table(tableName).CreateInBatches(datas, batchSize)
	if tx.Error != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlBatchCreate, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, tx.Error, "tableName", tableName, "datas", datas))
		return -1
	}
	return tx.RowsAffected
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
					zaplog.Errorw(LOG)
					tx.Rollback()
					return partwork.Error
				}
			} else {
				partwork := tx.Exec(sql)
				if partwork.Error != nil {
					zaplog.Errorw(LOG)
					tx.Rollback()
					return partwork.Error
				}
			}
		}
		if tx.Error != nil {
			zaplog.Errorw(LOG)
			tx.Rollback()
		}
		return tx.Commit().Error
	}
*/

// GORM自動transaction,sql raw執行,輸出生效行數
func (gormDB *GormDB) Transaction(traceId string, sqlStrings []string, params ...[]interface{}) (rowsAffect int64) {
	err := sqlDB.Transaction(func(tx *gorm.DB) error {
		//循序執行所有sql,累加rowsAffect,如果有錯誤gorm會rollback,正常結束會自動commit
		for i, sql := range sqlStrings {
			if params != nil {
				partwork := tx.Exec(sql, params[i]...)
				if partwork.Error != nil {
					zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, partwork.Error, "sql", sql, "params", params[i]))
					return partwork.Error
				}
				rowsAffect += partwork.RowsAffected
			} else {
				partwork := tx.Exec(sql)
				if partwork.Error != nil {
					zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.SqlTransaction, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, partwork.Error, "sql", sql))
					return partwork.Error
				}
				rowsAffect += partwork.RowsAffected
			}
		}

		return nil
	})
	//如果grom transaction有異常就返回無更新筆數,因為transaction已經log不再加log
	if err != nil {
		return -1
	}
	return rowsAffect
}
