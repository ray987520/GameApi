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
	defer tracer.PanicTrace(tracer.DefaultTraceId)

	initBaseServices()
	zaplog.DeleteIndex()
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(mconfig.GetString("application.listenPort"), routers)
}

func initBaseServices() {
	zaplog.InitZaplog()
	mconfig.InitConfigManager()
	es.InitCrypt()
	es.InitSonyflake()
	service.InitConcurrentService()
}
