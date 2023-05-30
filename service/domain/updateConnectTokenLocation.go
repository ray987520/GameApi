package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"io/ioutil"
	"net/http"
)

type UpdateTokenLocationService struct {
	Request  entity.UpdateTokenLocationRequest
	TraceMap string
}

// databinding&validate
func ParseUpdateTokenLocationRequest(traceMap string, r *http.Request) (request entity.UpdateTokenLocationRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
	}
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *UpdateTokenLocationService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	isUpdateOK := updateTokenLocation(es.AddTraceMap(service.TraceMap, string(functionid.UpdateTokenLocation)), &service.Request.BaseSelfDefine, service.Request.Token, service.Request.Location)
	if !isUpdateOK {
		return
	}
	return
}

// 更新connectToken location
func updateTokenLocation(traceMap string, selfDefine *entity.BaseSelfDefine, token string, location int) (isOK bool) {
	isOK = database.UpdateTokenLocation(es.AddTraceMap(traceMap, sqlid.UpdateTokenLocation.String()), token, location)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return
}

// 檢查aes token活耀
func isConnectTokenAlive(traceMap string, token string) bool {
	alive := database.GetTokenAlive(es.AddTraceMap(traceMap, sqlid.GetTokenAlive.String()), token)
	return alive
}

func (service *UpdateTokenLocationService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
