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

type SettlementService struct {
	Request  entity.SettlementRequest
	TraceMap string
}

// databinding&validate
func ParseSettlementRequest(traceMap string, r *http.Request) (request entity.SettlementRequest, err error) {
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

func (service *SettlementService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	isOK := addUnpayActivityRank(es.AddTraceMap(service.TraceMap, string(functionid.AddUnpayActivityRank)), &service.Request.BaseSelfDefine, service.Request.Settlement)
	if !isOK {
		return nil
	}

	return nil
}

// add未派彩紀錄
func addUnpayActivityRank(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Settlement) (isOK bool) {
	isOK = database.AddActivityRank(es.AddTraceMap(traceMap, sqlid.AddActivityRank.String()), data)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *SettlementService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
