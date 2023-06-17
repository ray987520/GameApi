package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/external/service/tracer"
	"TestAPI/service"
	"net/http"
)

// @Summary	取得單一注單將號2.1
// @Tags		BetSlip
// @Param		Authorization	header		string	true	"auth token"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/betSlip/getSequenceNumber [get]
func GetSequenceNumber(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.GetSequenceNumber, r)
	writeHttpResponse(w, traceId)
}

// @Summary	取得複數注單將號2.2
// @Tags		BetSlip
// @Param		Authorization	header		string	true	"auth token"
// @Param		quantity		query		int		true	"數量(可接受範圍1-50)"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/betSlip/getSequenceNumbers [get]
func GetSequenceNumbers(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.GetSequenceNumbers, r)
	writeHttpResponse(w, traceId)
}

// @Summary	取得需補注單列表2.3
// @Tags		BetSlip
// @Param		Authorization	header		string	true	"auth token"
// @Param		fromDate		query		string	true	"開始時間"
// @Param		toDate			query		string	true	"結束時間"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/betSlip/roundCheck [get]
func RoundCheck(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.RoundCheck, r)
	writeHttpResponse(w, traceId)
}
