package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/external/service/tracer"
	"net/http"
)

type CurrencyListService struct {
	Request entity.CurrencyListRequest
}

// databinding&validate
func ParseCurrencyListRequest(traceId string, r *http.Request) (request entity.CurrencyListRequest) {
	//read request header
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

func (service *CurrencyListService) Exec() interface{} {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	currencyList := getSupportCurrency(&service.Request.BaseSelfDefine)
	if currencyList == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return currencyList
}

// 取支援的Currency清單
func getSupportCurrency(selfDefine *entity.BaseSelfDefine) interface{} {
	currencyList := database.GetCurrencyList(selfDefine.TraceID)
	if currencyList == nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return currencyList
}

func (service *CurrencyListService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
