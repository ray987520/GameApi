package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	"TestAPI/service"
	"net/http"
)

//	@Summary	個人錢包轉至遊戲錢包5.1
//	@Tags		Transaction
//	@Param		Authorization	header		string				true	"auth token"
//	@Param		Body			body		entity.RollHistory	true	"body"
//	@Success	200				{object}	entity.BaseHttpResponse
//	@Router		/api/transaction/rollOut [post]
func RollOut(w http.ResponseWriter, r *http.Request) {
	traceId := initResponseChannel(r)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.RollOut), string(serviceid.ConcurrentEntry)), controllerid.RollOut, r)
	writeHttpResponse(w, traceId)
}

//	@Summary	遊戲錢包轉至個人錢包5.2
//	@Tags		Transaction
//	@Param		Authorization	header		string					true	"auth token"
//	@Param		Body			body		entity.RollInHistory	true	"body"
//	@Success	200				{object}	entity.BaseHttpResponse
//	@Router		/api/transaction/rollIn [post]
func RollIn(w http.ResponseWriter, r *http.Request) {
	traceId := initResponseChannel(r)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.RollIn), string(serviceid.ConcurrentEntry)), controllerid.RollIn, r)
	writeHttpResponse(w, traceId)
}
