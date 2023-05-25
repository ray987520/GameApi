package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"time"
)

const (
	ApiTimeFormat = "2006-01-02T15:04:05.999-07:00" //文件定義API時間格式
	DbTimeFormat  = "2006-01-02 15:04:05.999"       //sql db吃的時間格式
)

// UTC時間Timestamp
func Timestamp() int64 {
	return UtcNow().Unix()
}

// UTC時間
func UtcNow() time.Time {
	return time.Now().UTC()
}

// 本地時間,依輸入時區計算,例如LocalNow(8)=>+08:00時區
func LocalNow(areaPower int) time.Time {
	zone := time.FixedZone("", areaPower*60*60)
	return UtcNow().In(zone)
}

// 依2006-01-02T15:04:05.999-07:00轉出時間字串
func ApiTimeString(t time.Time) string {
	return t.Format(ApiTimeFormat)
}

// 按format parse時間
func ParseTime(traceMap, format, timeString string) (t time.Time, err error) {
	t, err = time.Parse(format, timeString)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ParseTime, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", format, "timeString", timeString)
		return
	}
	return
}
