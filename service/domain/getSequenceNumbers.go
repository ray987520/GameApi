package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/external/service/str"
	"TestAPI/external/service/tracer"
	"net/http"
)

type GetSequenceNumbersService struct {
	Request entity.GetSequenceNumbersRequest
}

// databinding&validate
func ParseGetSequenceNumbersRequest(traceId string, r *http.Request) (request entity.GetSequenceNumbersRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//read query string
	query := r.URL.Query()
	qty, isOK := str.Atoi(traceId, query.Get("quantity"))
	//convert error
	if !isOK {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	request.Quantity = qty

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	return request
}

func (service *GetSequenceNumbersService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//取多將號
	seqNos := getGameSequenceNumbers(&service.Request.BaseSelfDefine, service.Request.Quantity)
	if seqNos == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.GetSequenceNumbersResponse{
		SequenceNumber: seqNos,
	}
}

// 取多個將號,暫時prefix給空字串,如果redis數字爆掉可以加上新prefix避免重覆
func getGameSequenceNumbers(selfDefine *entity.BaseSelfDefine, qty int) []string {
	seqNos := database.GetGameSequenceNumbers(selfDefine.TraceID, qty, gameSequenceNumberPrefix)
	//get game sequence numbers error
	if len(seqNos) != qty {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}

	return seqNos
}

func (service *GetSequenceNumbersService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
