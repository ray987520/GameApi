package router

import (
	"TestAPI/controller"
	"TestAPI/enum/errorcode"
	"TestAPI/service"
	"net/http"

	_ "TestAPI/docs"

	esid "TestAPI/enum/externalserviceid"
	es "TestAPI/external/service"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// api router結構
type Route struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	Middlewares []mux.MiddlewareFunc
}

var (
	routes []Route
)

const (
	apitoken            = "999" //api auth token
	responseFormatError = "Http Response Json Format Error"
	authHeader          = "Authorization"
	traceHeader         = "traceid"
	requestTimeHeader   = "requesttime"
	errorCodeHeader     = "errorcode"
	swaggerPath         = "/swagger"
	apiPath             = "/api"
)

// 初始化,註冊所有api controller/middleware跟api path對應
func init() {
	register("GET", "/token/createGuestConnectToken", controller.CreateGuestConnectToken, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/v1.0/connectToken/authorization", controller.AuthConnectToken, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/token/updateConnectTokenLocation", controller.UpdateTokenLocation, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/token/getConnectTokenInfo", controller.GetConnectTokenInfo, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/token/getConnectTokenAmount", controller.GetConnectTokenAmount, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/token/delConnectToken", controller.DelConnectToken, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/betSlip/getSequenceNumber", controller.GetSequenceNumber, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/betSlip/getSequenceNumbers", controller.GetSequenceNumbers, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/betSlip/roundCheck", controller.RoundCheck, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/v1.0/betSlipPersonal/gameResult", controller.GameResult, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/v1.0/betSlipPersonal/supplement/result", controller.FinishGameResult, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/betSlipPersonal/addUniversalGameLog", controller.AddGameLog, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/gameReport/orderList", controller.OrderList, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/transaction/rollOut", controller.RollOut, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/transaction/rollIn", controller.RollIn, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/v1.0/activity/ranking/settlement", controller.Settlement, TraceIDMiddleware, AuthMiddleware)
	register("POST", "/v1.0/activity/ranking/distribution", controller.Distribution, TraceIDMiddleware, AuthMiddleware)
	register("GET", "/currency/currencyList", controller.CurrencyList, TraceIDMiddleware, AuthMiddleware)
}

// 使用mux Router,分不同前路徑規則劃分為swagger|api,使用不同middleware
func NewRouter() http.Handler {
	r := mux.NewRouter()
	r.PathPrefix(swaggerPath).Handler(httpSwagger.WrapHandler)
	apiRouter := r.PathPrefix(apiPath).Subrouter()
	for _, route := range routes {
		apiRouter.Methods(route.Method).
			Path(route.Pattern).
			Handler(route.Handler)
		if route.Middlewares != nil {
			for _, middleware := range route.Middlewares {
				apiRouter.Use(middleware)
			}
		}
	}
	handler := cors.Default().Handler(r)
	return handler
}

// 註冊url對controller映射及controller前端middlewares
func register(method, pattern string, handler http.HandlerFunc, middlewares ...mux.MiddlewareFunc) {
	routes = append(routes, Route{method, pattern, handler, middlewares})
}

/*TODO 確認取IP的方式後應增加IP白名單middleware
func IPWhiteListMiddleware
*/

// Auth Token驗證middleware
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get(authHeader)
		if token != apitoken {
			reqTime := req.Header.Get(requestTimeHeader)
			traceCode := req.Header.Get(traceHeader)
			response := service.GetHttpResponse(string(errorcode.BadParameter), reqTime, traceCode, "")
			data, err := es.JsonMarshal(traceCode, response)
			if err != nil {
				data = []byte(responseFormatError)
			}
			w.Write(data)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// 添加自訂資料middleware,主要有traceid/requesttime
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceID, err := es.Gen(es.AddTraceMap("", string(esid.UuidGen)))
		if err != nil {
			return
		}
		req.Header.Add(traceHeader, traceID)
		req.Header.Add(requestTimeHeader, es.ApiTimeString(es.LocalNow(8)))
		req.Header.Add(errorCodeHeader, string(errorcode.Success))
		next.ServeHTTP(w, req)
	})
}
