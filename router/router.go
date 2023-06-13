package router

import (
	"TestAPI/controller"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/innererror"
	"TestAPI/enum/middlewareid"
	"TestAPI/service"
	"fmt"
	"net/http"

	_ "TestAPI/docs"

	es "TestAPI/external/service"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"

	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/exp/slices"
)

// api router結構
type Route struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	Middlewares []mux.MiddlewareFunc
}

var (
	routes    []Route
	apiTokens = mconfig.GetStringSlice("application.apiToken") //api auth tokens
)

const (
	responseFormatError = "Http Response Json Format Error"
	genTraceIdError     = "Gen traceID Error"
	authHeader          = "Authorization"
	traceHeader         = "traceid"
	requestTimeHeader   = "requesttime"
	errorCodeHeader     = "errorcode"
	swaggerPath         = "/swagger"
	apiPath             = "/api"
	pprofPath           = "/debug"
	logRequest          = "Log Request"
	logResponse         = "Log Response"
	badErrorCode        = "no error code"
	authTokenError      = "the authtoken invalid"
)

// 初始化,註冊所有api controller/middleware跟api path對應
func init() {
	register("GET", "/token/createGuestConnectToken", controller.CreateGuestConnectToken, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/v1.0/connectToken/authorization", controller.AuthConnectToken, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/token/updateConnectTokenLocation", controller.UpdateTokenLocation, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/token/getConnectTokenInfo", controller.GetConnectTokenInfo, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/token/getConnectTokenAmount", controller.GetConnectTokenAmount, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/token/delConnectToken", controller.DelConnectToken, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/betSlip/getSequenceNumber", controller.GetSequenceNumber, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/betSlip/getSequenceNumbers", controller.GetSequenceNumbers, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/betSlip/roundCheck", controller.RoundCheck, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/v1.0/betSlipPersonal/gameResult", controller.GameResult, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/v1.0/betSlipPersonal/supplement/result", controller.FinishGameResult, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/betSlipPersonal/addUniversalGameLog", controller.AddGameLog, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/gameReport/orderList", controller.OrderList, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/transaction/rollOut", controller.RollOut, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/transaction/rollIn", controller.RollIn, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/v1.0/activity/ranking/settlement", controller.Settlement, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("POST", "/v1.0/activity/ranking/distribution", controller.Distribution, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
	register("GET", "/currency/currencyList", controller.CurrencyList, TraceIDMiddleware, AuthMiddleware, ErrorHandleMiddleware)
}

// init pprof web ui
func initPprofRouter(router *mux.Router) {
	router.Methods("GET").Path("/pprof").HandlerFunc(pprof.Index)
	router.Methods("GET").Path("/allocs").Handler(pprof.Handler("allocs"))
	router.Methods("GET").Path("/block").Handler(pprof.Handler("block"))
	router.Methods("GET").Path("/cmdline").HandlerFunc(pprof.Cmdline)
	router.Methods("GET").Path("/goroutine").Handler(pprof.Handler("goroutine"))
	router.Methods("GET").Path("/heap").Handler(pprof.Handler("heap"))
	router.Methods("GET").Path("/mutex").Handler(pprof.Handler("mutex"))
	router.Methods("GET").Path("/profile").HandlerFunc(pprof.Profile)
	router.Methods("GET").Path("/threadcreate").Handler(pprof.Handler("threadcreate"))
	router.Methods("GET").Path("/trace").HandlerFunc(pprof.Trace)
}

// 使用mux Router,分不同前路徑規則劃分為swagger|api,使用不同middleware
func NewRouter() http.Handler {
	r := mux.NewRouter()
	//swagger走自己的路徑不用經過middleware,/swagger,default寫法:r.PathPrefix(swaggerPath).Handler(httpSwagger.WrapHandler)
	//部分ui畫面可以自訂的寫法如下,可以控制有沒有swagger外框,插入plugin/uiconfig的JS
	r.PathPrefix(swaggerPath).Handler(httpSwagger.Handler(httpSwagger.Layout(httpSwagger.StandaloneLayout), httpSwagger.DeepLinking(true)))
	//pprof走自己的路徑不用經過middleware,/debug
	pprofRouter := r.PathPrefix(pprofPath).Subrouter()
	initPprofRouter(pprofRouter)
	//api走subrouter加上middleware
	apiRouter := r.PathPrefix(apiPath).Subrouter()
	for _, route := range routes {
		apiSubRouter := apiRouter.Methods(route.Method).Path(route.Pattern).Subrouter()
		apiSubRouter.NewRoute().Handler(route.Handler)
		apiSubRouter.Use(route.Middlewares...)
	}
	handler := cors.Default().Handler(r)
	return handler
}

// 註冊url對controller映射及controller前端middlewares
func register(method, pattern string, handler http.HandlerFunc, middlewares ...mux.MiddlewareFunc) {
	routes = append(routes, Route{method, pattern, handler, middlewares})
}

/**TODO 確認取IP的方式後應增加IP白名單middleware
func IPWhiteListMiddleware
*/

// Auth Token驗證middleware
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get(authHeader)
		//如果白名單auth token不包含輸入token就返回驗證失敗,GO1.18支援原生判斷slice contains
		if !slices.Contains(apiTokens, token) {
			reqTime := req.Header.Get(requestTimeHeader)
			traceID := req.Header.Get(traceHeader)
			response := service.GetHttpResponse(string(errorcode.BadParameter), reqTime, traceID, "")
			data := es.JsonMarshal(traceID, response)
			if data == nil {
				data = []byte(responseFormatError)
			}
			w.Write(data)
			err := fmt.Errorf(authTokenError)
			zaplog.Errorw(innererror.MiddlewareError, innererror.FunctionNode, middlewareid.AuthMiddleware, innererror.TraceNode, traceID, innererror.ErrorInfoNode, err, "token", token)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// 添加自訂資料middleware,主要有traceid/requesttime
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceID := es.Gen(tracer.DefaultTraceId)
		//唯一的traceID產生失敗就返回異常
		if traceID == "" {
			reqTime := es.ApiTimeString(es.LocalNow(8))
			response := service.GetHttpResponse(string(errorcode.UnknowError), reqTime, traceID, "")
			data := es.JsonMarshal(traceID, response)
			if data == nil {
				data = []byte(responseFormatError)
			}
			w.Write(data)
			err := fmt.Errorf(genTraceIdError)
			zaplog.Errorw(innererror.MiddlewareError, innererror.FunctionNode, middlewareid.TraceIDMiddleware, innererror.TraceNode, tracer.DefaultTraceId, innererror.ErrorInfoNode, err, "traceID", traceID)
			return
		}
		//記錄原始http request
		logOriginRequest(req, traceID)
		req.Header.Add(traceHeader, traceID)
		req.Header.Add(requestTimeHeader, es.ApiTimeString(es.LocalNow(8)))
		req.Header.Add(errorCodeHeader, string(errorcode.Default))
		next.ServeHTTP(w, req)
	})
}

// 封裝error中間件,在跳過正常的ResponseWriter時,仍封裝輸出標準錯誤
func ErrorHandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req)
		traceID := req.Header.Get(traceHeader)
		zaplog.Infow(logResponse, innererror.FunctionNode, middlewareid.ErrorHandleMiddleware, innererror.TraceNode, traceID, "responseHeaders", w.Header())
		//在response之後檢查
		errorCode := req.Header.Get(errorCodeHeader)
		//正常的話errorCode應為0或其他值,空值代表異常結束
		if errorCode == "" {
			reqTime := req.Header.Get(requestTimeHeader)
			response := service.GetHttpResponse(string(errorcode.UnknowError), reqTime, traceID, "")
			data := es.JsonMarshal(traceID, response)
			if data == nil {
				data = []byte(responseFormatError)
			}
			w.Write(data)
			err := fmt.Errorf(badErrorCode)
			zaplog.Errorw(innererror.MiddlewareError, innererror.FunctionNode, middlewareid.ErrorHandleMiddleware, innererror.TraceNode, traceID, innererror.ErrorInfoNode, err, "errorCode", errorCode)
			return
		}

	})
}

// 記錄原始http request
func logOriginRequest(req *http.Request, traceId string) {
	curl, err := service.HttpRequest2Curl(req)
	zaplog.Infow(logRequest, innererror.FunctionNode, middlewareid.LogOriginRequest, innererror.TraceNode, traceId, "curl", curl, innererror.ErrorInfoNode, err)
}
