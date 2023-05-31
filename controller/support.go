package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	"TestAPI/service"
	"net/http"
)

// @Summary	取得支援幣別7.1
// @Tags		Support
// @Param		Authorization	header		string	true	"auth token"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/currency/currencyList [get]
func CurrencyList(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer es.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.CurrencyList), string(serviceid.ConcurrentEntry)), controllerid.CurrencyList, r)
	writeHttpResponse(w, traceId)
}
