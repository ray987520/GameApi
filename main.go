package main

import (
	_ "TestAPI/docs"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/router"
	"net/http"
)

func main() {
	defer tracer.PanicTrace(tracer.DefaultTraceId)
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(mconfig.GetString("application.listenPort"), routers)
}
