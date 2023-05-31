package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	"TestAPI/service"
	"net/http"
)

// @Summary	寫入賽果(拉霸)3.1
// @Tags		BetSlipPersonal
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.GameResult	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/betSlipPersonal/gameResult [post]
func GameResult(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer es.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.GameResult), string(serviceid.ConcurrentEntry)), controllerid.GameResult, r)
	writeHttpResponse(w, traceId)
}

// @Summary	補單補賽果單3.2
// @Tags		BetSlipPersonal
// @Param		Authorization	header		string					true	"auth token"
// @Param		Body			body		entity.FinishGameResult	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/betSlipPersonal/supplement/result [post]
func FinishGameResult(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer es.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.FinishGameResult), string(serviceid.ConcurrentEntry)), controllerid.FinishGameResult, r)
	writeHttpResponse(w, traceId)
}

// @Summary	寫入遊戲紀錄3.3
// @Tags		BetSlipPersonal
// @Param		Authorization	header		string			true	"auth token"
// @Param		Body			body		entity.GameLog	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/betSlipPersonal/addUniversalGameLog [post]
func AddGameLog(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer es.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.AddGameLog), string(serviceid.ConcurrentEntry)), controllerid.AddGameLog, r)
	writeHttpResponse(w, traceId)
}
