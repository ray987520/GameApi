package mconfig

import (
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	configFileName       = "config"
	viperReadFileError   = "viper read config file error:%v"
	viperReadConfigError = "viper read config error ,configPath:%s ,data:%v"
	configChangeMessage  = "config file changed ,data:%s"
)

// 初始化viper
func InitConfigManager() {
	viper.AddConfigPath("./")
	viper.SetConfigName(configFileName)
	err := viper.ReadInConfig()
	//讀取設定檔失敗,不繼續執行代碼
	if err != nil {
		err = fmt.Errorf(viperReadFileError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, innererror.MConfigInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, err)
		panic(err)
	}
	//動態監看設定檔更新
	viper.WatchConfig()
	//設定檔更新時列印更新欄位
	viper.OnConfigChange(func(e fsnotify.Event) {
		msg := fmt.Sprintf(configChangeMessage, e.Name)
		zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, innererror.MConfigInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, msg)
	})
}

// 取設定值,string
func GetString(configPath string) string {
	return cast.ToString(Get(configPath))
}

// 取設定值,int
func GetInt(configPath string) int {
	return cast.ToInt(Get(configPath))
}

// 取設定值,int64
func GetInt64(configPath string) int64 {
	return cast.ToInt64(Get(configPath))
}

// 取設定值,時間區間
func GetDuration(configPath string) time.Duration {
	return cast.ToDuration(Get(configPath))
}

// 取設定值,[]string
func GetStringSlice(configPath string) []string {
	return cast.ToStringSlice(Get(configPath))
}

// 取設定值,interface{}
func Get(configPath string) any {
	data := viper.Get(configPath)
	//如果找不到設定值,不希望代碼繼續執行
	if data == nil {
		err := fmt.Errorf(viperReadConfigError, configPath, data)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, innererror.MConfigGet, innererror.TraceNode, tracer.DefaultTraceId, innererror.DataNode, err)
		panic(err)
	}
	return data
}
