package es

import (
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
func TimeString(t time.Time) string {
	return t.Format(ApiTimeFormat)
}
