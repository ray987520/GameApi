package controller

import (
	"TestAPI/enum/controllerid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	"TestAPI/service"
	"net/http"
)

//	@Summary	活動結算6.1
//	@Tags		Activity
//	@Param		Authorization	header		string				true	"auth token"
//	@Param		Body			body		entity.Settlement	true	"body"
//	@Success	200				{object}	entity.BaseHttpResponse
//	@Router		/api/v1.0/activity/ranking/settlement [post]
func Settlement(w http.ResponseWriter, r *http.Request) {
	traceId := initResponseChannel(r)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.Settlement), string(serviceid.ConcurrentEntry)), controllerid.Settlement, r)
	writeHttpResponse(w, traceId)
}

//	@Summary	活動派獎6.2
//	@Tags		Activity
//	@Param		Authorization	header		string				true	"auth token"
//	@Param		Body			body		entity.Distribution	true	"body"
//	@Success	200				{object}	entity.BaseHttpResponse
//	@Router		/api/v1.0/activity/ranking/distribution [post]
func Distribution(w http.ResponseWriter, r *http.Request) {
	traceId := initResponseChannel(r)
	service.Entry(es.AddTraceMap(traceId+"_"+string(controllerid.Distribution), string(serviceid.ConcurrentEntry)), controllerid.Distribution, r)
	writeHttpResponse(w, traceId)
}
