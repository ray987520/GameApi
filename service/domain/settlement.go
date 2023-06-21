package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"net/http"
)

type SettlementService struct {
	Request entity.SettlementRequest
}

// databinding&validate
func ParseSettlementRequest(traceId string, r *http.Request) (request entity.SettlementRequest) {
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

	//read request
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *SettlementService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//add活動待派彩資料
	isOK := addUnpayActivityRank(&service.Request.BaseSelfDefine, service.Request.Settlement)
	if !isOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// add未派彩紀錄
func addUnpayActivityRank(selfDefine *entity.BaseSelfDefine, data entity.Settlement) (isOK bool) {
	isOK = database.AddActivityRank(selfDefine.TraceID, data)
	//add record error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}

	return isOK
}

func (service *SettlementService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
