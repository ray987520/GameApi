package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
	"net/http"
	"strconv"
)

type GetSequenceNumbersService struct {
	Request  entity.GetSequenceNumbersRequest
	TraceMap string
}

// databinding&validate
func ParseGetSequenceNumbersRequest(traceMap string, r *http.Request) (request entity.GetSequenceNumbersRequest, err error) {
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	query := r.URL.Query()
	qty, err := strconv.Atoi(query.Get("quantity"))
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	request.Quantity = qty
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GetSequenceNumbersService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	seqNos := getGameSequenceNumbers(es.AddTraceMap(service.TraceMap, string(functionid.GetGameSequenceNumbers)), &service.Request.BaseSelfDefine, service.Request.Quantity)
	if seqNos == nil {
		return
	}
	data = entity.GetSequenceNumbersResponse{
		SequenceNumber: seqNos,
	}
	return
}

// 取多個將號,暫時prefix給空字串,如果redis數字爆掉可以加上新prefix避免重覆
func getGameSequenceNumbers(traceMap string, selfDefine *entity.BaseSelfDefine, qty int) []string {
	seqNos := database.GetGameSequenceNumbers(es.AddTraceMap(traceMap, redisid.GetGameSequenceNumbers.String()), qty, "")
	if len(seqNos) != qty {
		es.Error("traceMap:%s ,getGameSequenceNumbers error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return seqNos
}

func (service *GetSequenceNumbersService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
