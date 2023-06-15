package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"net/http"
)

type UpdateTokenLocationService struct {
	Request  entity.UpdateTokenLocationRequest
	TraceMap string
}

// databinding&validate
func ParseUpdateTokenLocationRequest(traceMap string, r *http.Request) (request entity.UpdateTokenLocationRequest, err error) {
	body, err := readHttpRequestBody(es.AddTraceMap(traceMap, string(functionid.ReadHttpRequestBody)), r, &request)
	if err != nil {
		return request, err
	}

	err = parseJsonBody(es.AddTraceMap(traceMap, string(functionid.ParseJsonBody)), body, &request)
	if err != nil {
		return request, err
	}

	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *UpdateTokenLocationService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	isUpdateOK := updateTokenLocation(es.AddTraceMap(service.TraceMap, string(functionid.UpdateTokenLocation)), &service.Request.BaseSelfDefine, service.Request.Token, service.Request.Location)
	if !isUpdateOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 更新connectToken location
func updateTokenLocation(traceMap string, selfDefine *entity.BaseSelfDefine, token string, location int) (isOK bool) {
	isOK = database.UpdateTokenLocation(es.AddTraceMap(traceMap, sqlid.UpdateTokenLocation.String()), token, location)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 檢查aes token活耀
func isConnectTokenAlive(traceMap string, token string) bool {
	alive := database.GetTokenAlive(es.AddTraceMap(traceMap, sqlid.GetTokenAlive.String()), token)
	return alive
}

func (service *UpdateTokenLocationService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
