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

type CurrencyListService struct {
	Request  entity.CurrencyListRequest
	TraceMap string
}

// databinding&validate
func ParseCurrencyListRequest(traceMap string, r *http.Request) (request entity.CurrencyListRequest, err error) {
	//read request header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//validate request
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *CurrencyListService) Exec() interface{} {
	//catch panic
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	data := getSupportCurrency(es.AddTraceMap(service.TraceMap, string(functionid.GetSupportCurrency)), &service.Request.BaseSelfDefine)
	return data
}

// 取支援的Currency清單
func getSupportCurrency(traceMap string, selfDefine *entity.BaseSelfDefine) interface{} {
	currencyList, err := database.GetCurrencyList(es.AddTraceMap(traceMap, sqlid.GetCurrencyList.String()))
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return currencyList
}

func (service *CurrencyListService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
