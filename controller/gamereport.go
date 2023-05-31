package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	"TestAPI/service"
	"net/http"
)

// @Summary	取得歷史紀錄網址4.1
// @Tags		GameReport
// @Param		Authorization	header		string	true	"auth token"
// @Param		connectToken	query		string	true	"連線令牌"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/gameReport/orderList [get]
func OrderList(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer es.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.OrderList), string(serviceid.ConcurrentEntry)), controllerid.OrderList, r)
	writeHttpResponse(w, traceId)
}
