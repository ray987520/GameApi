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

type SettlementService struct {
	Request  entity.SettlementRequest
	TraceMap string
}

// databinding&validate
func ParseSettlementRequest(traceMap string, r *http.Request) (request entity.SettlementRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
	}
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

func (service *SettlementService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	isOK := addUnpayActivityRank(es.AddTraceMap(service.TraceMap, string(functionid.AddUnpayActivityRank)), &service.Request.BaseSelfDefine, service.Request.Settlement)
	if !isOK {
		return
	}
	return
}

// add未派彩紀錄
func addUnpayActivityRank(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Settlement) (isOK bool) {
	isOK = database.AddActivityRank(es.AddTraceMap(traceMap, sqlid.AddActivityRank.String()), data)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *SettlementService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
