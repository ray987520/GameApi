package main

import (
	_ "TestAPI/docs"
	"TestAPI/router"
	"net/http"
)

func main() {
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(":8080", routers)
}
