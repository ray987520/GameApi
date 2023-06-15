package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"net/http"
)

type DelConnectTokenService struct {
	Request entity.DelConnectTokenRequest
}

// databinding&validate
func ParseDelConnectTokenRequest(traceId string, r *http.Request) (request entity.DelConnectTokenRequest) {
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

func (service *DelConnectTokenService) Exec() interface{} {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	isOK := logOutToken(&service.Request.BaseSelfDefine, service.Request.Token)
	if !isOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// remove token cache and set token deleted
func logOutToken(selfDefine *entity.BaseSelfDefine, token string) bool {
	now := es.LocalNow(8)
	isOK := database.DeleteToken(selfDefine.TraceID, token, now)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *DelConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
