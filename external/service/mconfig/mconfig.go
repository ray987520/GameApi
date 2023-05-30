package mconfig

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	configFileName       = "config"
	viperReadFileError   = "Viper Read Config File Error:%v"
	viperReadConfigError = "Viper Read Config Error ,configPath:%s ,data:%v"
	configChangeMessage  = "Config File Changed ,data:%s"
)

func init() {
	viper.AddConfigPath("./")
	viper.SetConfigName(configFileName)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf(viperReadFileError, err))
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println(fmt.Sprintf(configChangeMessage, e.Name))
	})
}

func GetString(configPath string) string {
	return cast.ToString(Get(configPath))
}

func GetInt(configPath string) int {
	return cast.ToInt(Get(configPath))
}

func GetInt64(configPath string) int64 {
	return cast.ToInt64(Get(configPath))
}

func GetDuration(configPath string) time.Duration {
	return cast.ToDuration(Get(configPath))
}

func Get(configPath string) any {
	data := viper.Get(configPath)
	if data == nil {
		err := fmt.Errorf(viperReadConfigError, configPath, data)
		fmt.Println(err)
		panic(err)
	}
	return data
}
