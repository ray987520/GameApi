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

type DelConnectTokenService struct {
	Request  entity.DelConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseDelConnectTokenRequest(traceMap string, r *http.Request) (request entity.DelConnectTokenRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
		return
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
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

func (service *DelConnectTokenService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	logOutToken(es.AddTraceMap(service.TraceMap, string(functionid.LogOutToken)), &service.Request.BaseSelfDefine, service.Request.Token)
	return
}

func logOutToken(traceMap string, selfDefine *entity.BaseSelfDefine, token string) (isOK bool) {
	now := es.LocalNow(8)
	isOK = database.DeleteToken(es.AddTraceMap(traceMap, sqlid.DeleteToken.String()), token, now)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *DelConnectTokenService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
