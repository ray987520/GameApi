package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"net/http"
	"strconv"
)

type GetSequenceNumbersService struct {
	Request  entity.GetSequenceNumbersRequest
	TraceMap string
}

// databinding&validate
func ParseGetSequenceNumbersRequest(traceMap string, r *http.Request) (request entity.GetSequenceNumbersRequest, err error) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	query := r.URL.Query()
	qty, err := strconv.Atoi(query.Get("quantity"))
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	request.Quantity = qty

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *GetSequenceNumbersService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	seqNos := getGameSequenceNumbers(es.AddTraceMap(service.TraceMap, string(functionid.GetGameSequenceNumbers)), &service.Request.BaseSelfDefine, service.Request.Quantity)
	if seqNos == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.GetSequenceNumbersResponse{
		SequenceNumber: seqNos,
	}
}

// 取多個將號,暫時prefix給空字串,如果redis數字爆掉可以加上新prefix避免重覆
func getGameSequenceNumbers(traceMap string, selfDefine *entity.BaseSelfDefine, qty int) []string {
	seqNos := database.GetGameSequenceNumbers(es.AddTraceMap(traceMap, redisid.GetGameSequenceNumbers.String()), qty, "")
	if len(seqNos) != qty {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return seqNos
}

func (service *GetSequenceNumbersService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
