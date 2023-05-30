package main

import (
	_ "TestAPI/docs"
	es "TestAPI/external/service"
	"TestAPI/external/service/mconfig"
	"TestAPI/router"
	"net/http"
)

func main() {
	defer es.PanicTrace("")
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(mconfig.GetString("application.listenPort"), routers)
}
