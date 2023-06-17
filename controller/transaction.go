package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/external/service/tracer"
	"TestAPI/service"
	"net/http"
)

// @Summary	個人錢包轉至遊戲錢包5.1
// @Tags		Transaction
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.RollHistory	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/transaction/rollOut [post]
func RollOut(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.RollOut, r)
	writeHttpResponse(w, traceId)
}

// @Summary	遊戲錢包轉至個人錢包5.2
// @Tags		Transaction
// @Param		Authorization	header		string					true	"auth token"
// @Param		Body			body		entity.RollInHistory	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/transaction/rollIn [post]
func RollIn(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.RollIn, r)
	writeHttpResponse(w, traceId)
}
