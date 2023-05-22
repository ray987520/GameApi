package main

import (
	_ "TestAPI/docs"
	"TestAPI/router"
	"net/http"
)

func main() {
	/*測試鎖
	sql := []string{`SELECT 1 FROM ErrorMessage WITH(HOLDLOCK) WHERE id=1`, `UPDATE ErrorMessage SET codeType=1 WHERE id=1`}
	es.GetSqlDb().Transaction(sql)
	*/
	//初始化api router,然後聆聽
	routers := router.NewRouter()
	http.ListenAndServe(":8080", routers)
}
