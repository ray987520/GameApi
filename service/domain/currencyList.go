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
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *CurrencyListService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	data = getSupportCurrency(es.AddTraceMap(service.TraceMap, string(functionid.GetSupportCurrency)), &service.Request.BaseSelfDefine)
	return
}

// 取支援的Currency清單
func getSupportCurrency(traceMap string, selfDefine *entity.BaseSelfDefine) interface{} {
	currencyList, err := database.GetCurrencyList(es.AddTraceMap(traceMap, sqlid.GetCurrencyList.String()))
	if err != nil {
		es.Error("traceMap:%s , error:%v", traceMap, err)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return currencyList
}

func (service *CurrencyListService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
