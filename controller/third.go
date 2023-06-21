package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/external/service/tracer"
	"TestAPI/service"
	"net/http"
)

// @Summary	踢除令牌 third1.1
// @Tags		Third
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.KickToken	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/apiControl/kickToken [post]
func KickToken(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.KickToken, r)
	writeHttpResponse(w, traceId)
}

// @Summary	確認令牌連線狀態 third1.2
// @Tags		Third
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.IsTokenOnline	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/apiControl/isTokenOnline [post]
func IsTokenOnline(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.IsTokenOnline, r)
	writeHttpResponse(w, traceId)
}
