package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/external/service/tracer"
	"TestAPI/service"
	"net/http"
)

// @Summary	活動結算6.1
// @Tags		Activity
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.Settlement	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/activity/ranking/settlement [post]
func Settlement(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.Settlement, r)
	writeHttpResponse(w, traceId)
}

// @Summary	活動派獎6.2
// @Tags		Activity
// @Param		Authorization	header		string				true	"auth token"
// @Param		Body			body		entity.Distribution	true	"body"
// @Success	200				{object}	entity.BaseHttpResponse
// @Router		/api/v1.0/activity/ranking/distribution [post]
func Distribution(w http.ResponseWriter, r *http.Request) {
	traceId := getTraceIdFromRequest(r)
	defer tracer.PanicTrace(traceId)
	initResponseChannel(traceId)
	service.Entry(traceId, controllerid.Distribution, r)
	writeHttpResponse(w, traceId)
}
