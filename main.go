package main

import (
	_ "TestAPI/docs"
	es "TestAPI/external/service"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"TestAPI/router"
	"TestAPI/service"
	"net/http"
)

func main() {
	//zaplog.DeleteIndex()
	defer tracer.PanicTrace(tracer.DefaultTraceId)

	//初始化底層服務
	initBaseServices()
	zaplog.DeleteIndex()
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(mconfig.GetString("application.listenPort"), routers)
}

// 初始化底層服務,原本散落在init()跟變數裡面,不好掌控且可能發生底層沒初始化就先被呼叫
func initBaseServices() {
	//初始化log底層,包含zap/els
	zaplog.InitZaplog()
	//初始化封裝的viper
	mconfig.InitConfigManager()
	//初始化加密服務
	es.InitCrypt()
	//初始化SONY的雪花ID
	es.InitSonyflake()
	//初始化併發核心
	service.InitConcurrentService()
}
