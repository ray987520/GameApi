package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

const (
	sonyFlakeBaseTime = "2023-01-01 00:00:00.000" //需要設置一個固定時間起點讓sonyFlakeID的timestamp區段不重複
)

var sonyFlake *sonyflake.Sonyflake

// 取機器ID
func getMachineID() (machineID uint16, err error) {
	//*TODO 暫時使用一個假的machineID,後續應有環境變數或其他方式提供機器ID
	machineID = 1688
	return
}

// 初始化,設置sonyFlake基礎值
func init() {
	beginTime, err := time.Parse(DbTimeFormat, sonyFlakeBaseTime)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.ErrorInfoNode, err)
		return
	}
	st := sonyflake.Settings{
		StartTime: beginTime,
	}
	st.MachineID = getMachineID
	sonyFlake = sonyflake.NewSonyflake(st)
}

// 產生sonyFlakeID
func Gen(traceMap string) (uuid string, err error) {
	if sonyFlake == nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.ErrorTypeNode, innererror.InitFlakeError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
		return "", err
	}
	id, err := sonyFlake.NextID()
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.ErrorTypeNode, innererror.GenUidError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
		return "", err
	}
	uuid = strconv.FormatUint(id, 16)
	return
}
