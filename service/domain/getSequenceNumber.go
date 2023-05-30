package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
	"net/http"
)

type GetSequenceNumberService struct {
	Request  entity.GetSequenceNumberRequest
	TraceMap string
}

// databinding&validate
func ParseGetSequenceNumberRequest(traceMap string, r *http.Request) (request entity.GetSequenceNumberRequest, err error) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	return
}

func (service *GetSequenceNumberService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
	if service.Request.HasError() {
		return
	}
	seqNo := getGameSequenceNumber(es.AddTraceMap(service.TraceMap, string(functionid.GetGameSequenceNumber)), &service.Request.BaseSelfDefine)
	if seqNo == "" {
		return
	}
	data = entity.GetSequenceNumberResponse{
		SequenceNumber: seqNo,
	}
	return
}

// 取單一將號,暫時prefix給空字串,如果redis數字爆掉可以加上新prefix避免重覆
func getGameSequenceNumber(traceMap string, selfDefine *entity.BaseSelfDefine) string {
	seqNo := database.GetGameSequenceNumber(es.AddTraceMap(traceMap, redisid.GetGameSequenceNumber.String()), "")
	if seqNo == "" {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return seqNo
}

func (service *GetSequenceNumberService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
