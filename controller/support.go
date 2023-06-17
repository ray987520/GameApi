package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/external/service/tracer"
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
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.CurrencyList, r)
	writeHttpResponse(w, traceId)
}
