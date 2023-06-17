package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/external/service/tracer"
	"net/http"
)

type UpdateTokenLocationService struct {
	Request entity.UpdateTokenLocationRequest
}

// databinding&validate
func ParseUpdateTokenLocationRequest(traceId string, r *http.Request) (request entity.UpdateTokenLocationRequest) {
	body, isOK := readHttpRequestBody(traceId, r, &request)
	//read body error
	if !isOK {
		return request
	}

	isOK = parseJsonBody(traceId, body, &request)
	//json deserialize error
	if !isOK {
		return request
	}

	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	return request
}

func (service *UpdateTokenLocationService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//update game token location
	isUpdateOK := updateTokenLocation(&service.Request.BaseSelfDefine, service.Request.Token, service.Request.Location)
	if !isUpdateOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 更新connectToken location
func updateTokenLocation(selfDefine *entity.BaseSelfDefine, token string, location int) (isOK bool) {
	isOK = database.UpdateTokenLocation(selfDefine.TraceID, token, location)
	//update location error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}

	return isOK
}

// 檢查aes token活耀
func isConnectTokenAlive(traceId string, token string) bool {
	alive := database.GetTokenAlive(traceId, token)
	return alive
}

func (service *UpdateTokenLocationService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
