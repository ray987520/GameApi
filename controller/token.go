package controller

import (
	"TestAPI/entity"
	"TestAPI/enum/controllerid"
	es "TestAPI/external/service"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/service"
	"net/http"

	"github.com/shopspring/decimal"
)

const (
	loadResponseChannelError = "Load Http Response Channel Error"
	responseFormatError      = "Http Response Json Format Error"
)

var (
	traceIdFieldName = mconfig.GetString("trace.idFieldName")
)

// @Summary	取得測試令牌1.0
// @Tags		Token
// @Param		Authorization	header		string	true	"auth token"
// @Param		account			query		string	true	"會員帳號"
// @Param		currency		query		string	true	"幣別 "
// @Param		gameID			query		string	true	"遊戲代碼"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/token/createGuestConnectToken [get]
func CreateGuestConnectToken(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.CreateGuestConnectToken, r)
	writeHttpResponse(w, traceId)
}

func getTraceIdFromRequest(r *http.Request) (traceID string) {
	return r.Header.Get(traceIdFieldName)
}

// 在公用MAP註冊一個traceid(uuid)的唯一response channel
func initResponseChannel(traceId string) {
	service.ResponseMap.Store(traceId, make(chan entity.BaseHttpResponse))
	//service.ResponseMap[traceID] = make(chan entity.BaseHttpResponse)
}

// response回寫到channel公用map
func writeHttpResponse(w http.ResponseWriter, traceId string) {
	//設定decimal序列化處理時不要雙引號
	decimal.MarshalJSONWithoutQuotes = true

	var (
		data []byte
	)

	//sync.Map不能用舊的map[key]方式取值賦值,改用sync.Map.Load取值
	value, isOK := service.ResponseMap.Load(traceId)
	if !isOK {
		//map找不到traceID的情況報錯
		data = []byte(loadResponseChannelError)
	} else {
		//先型別斷言responseChannel再取出response=>關閉responseChannel=>刪除map key
		responseChannel := value.(chan entity.BaseHttpResponse)
		response := <-responseChannel
		close(responseChannel)
		service.ResponseMap.Delete(traceId)
		//response := <-service.ResponseMap[traceID]
		//close(service.ResponseMap[traceID])
		//delete(service.ResponseMap, traceID)
		//序列化response
		data = es.JsonMarshal(traceId, response)
		//json serialize error
		if data == nil {
			data = []byte(responseFormatError)
		}
	}
	w.Write(data)
}

// @Summary	令牌登入1.1
// @Tags		Token
// @Param		Authorization	header		string					true	"auth token"
// @Param		Body			body		entity.AuthConnectToken	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/connectToken/authorization [post]
func AuthConnectToken(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.AuthConnectToken, r)
	writeHttpResponse(w, traceId)
}

// @Summary	變更令牌位置1.2
// @Tags		Token
// @Param		Authorization	header		string						true	"auth token"
// @Param		Body			body		entity.UpdateTokenLocation	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/token/updateConnectTokenLocation [post]
func UpdateTokenLocation(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.UpdateTokenLocation, r)
	writeHttpResponse(w, traceId)
}

// @Summary	取得令牌資訊1.3
// @Tags		Token
// @Param		Authorization	header		string	true	"auth token"
// @Param		connectToken	query		string	true	"連線令牌"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/token/getConnectTokenInfo [get]
func GetConnectTokenInfo(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.GetConnectTokenInfo, r)
	writeHttpResponse(w, traceId)
}

// @Summary	取得令牌餘額1.4
// @Tags		Token
// @Param		Authorization	header		string	true	"auth token"
// @Param		connectToken	query		string	true	"連線令牌"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/token/getConnectTokenAmount [get]
func GetConnectTokenAmount(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.GetConnectTokenAmount, r)
	writeHttpResponse(w, traceId)
}

// @Summary	令牌登出1.5
// @Tags		Token
// @Param		Authorization	header		string						true	"auth token"
// @Param		Body			body		entity.DelConnectTokenBody	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/token/delConnectToken [post]
func DelConnectToken(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.DelConnectToken, r)
	writeHttpResponse(w, traceId)
}
