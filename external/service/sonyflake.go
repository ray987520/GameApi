package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"fmt"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

const (
	sonyFlakeBaseTime  = "2023-01-01 00:00:00.000" //需要設置一個固定時間起點讓sonyFlakeID的timestamp區段不重複
	initFlakeTimeError = "init sonyflake base time error:%v"
	flakeInstanceError = "sonyflake instance error"
)

var sonyFlake *sonyflake.Sonyflake

// 取機器ID
func getMachineID() (machineID uint16, err error) {
	//*TODO 暫時使用一個假的machineID,後續應有環境變數或其他方式提供機器ID
	machineID = 1688
	return machineID, nil
}

// 初始化,設置sonyFlake基礎值
func InitSonyflake() {
	beginTime, err := time.Parse(DbTimeFormat, sonyFlakeBaseTime)
	if err != nil {
		err = fmt.Errorf(initFlakeTimeError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, tracer.DefaultTraceId, innererror.ErrorInfoNode, err)
		return
	}
	st := sonyflake.Settings{
		StartTime: beginTime,
	}
	st.MachineID = getMachineID
	sonyFlake = sonyflake.NewSonyflake(st)
}

// 產生sonyFlakeID
func Gen(traceId string) (uuid string) {
	if sonyFlake == nil {
		err := fmt.Errorf(flakeInstanceError)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return ""
	}
	id, err := sonyFlake.NextID()
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return ""
	}
	//轉成16進位數字字串(比較短)
	uuid = strconv.FormatUint(id, 16)
	return uuid
}
