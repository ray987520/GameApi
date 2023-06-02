package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"net/http"
)

type DelConnectTokenService struct {
	Request  entity.DelConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseDelConnectTokenRequest(traceMap string, r *http.Request) (request entity.DelConnectTokenRequest, err error) {
	body, err := readHttpRequestBody(es.AddTraceMap(traceMap, string(functionid.ReadHttpRequestBody)), r, &request)
	if err != nil {
		return
	}

	err = parseJsonBody(es.AddTraceMap(traceMap, string(functionid.ParseJsonBody)), body, &request)
	if err != nil {
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

func (service *DelConnectTokenService) Exec() interface{} {
	//catch panic
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	logOutToken(es.AddTraceMap(service.TraceMap, string(functionid.LogOutToken)), &service.Request.BaseSelfDefine, service.Request.Token)
	return nil
}

// remove token cache and set token deleted
func logOutToken(traceMap string, selfDefine *entity.BaseSelfDefine, token string) bool {
	now := es.LocalNow(8)
	isOK := database.DeleteToken(es.AddTraceMap(traceMap, sqlid.DeleteToken.String()), token, now)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return isOK
	}
	return isOK
}

func (service *DelConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
