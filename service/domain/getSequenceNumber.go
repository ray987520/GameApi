package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"net/http"
)

type GetSequenceNumberService struct {
	Request entity.GetSequenceNumberRequest
}

// databinding&validate
func ParseGetSequenceNumberRequest(traceId string, r *http.Request) (request entity.GetSequenceNumberRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	return request
}

func (service *GetSequenceNumberService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//取將號
	seqNo := getGameSequenceNumber(&service.Request.BaseSelfDefine)
	if seqNo == "" {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.GetSequenceNumberResponse{
		SequenceNumber: seqNo,
	}
}

// 取單一將號,暫時prefix給空字串,如果redis數字爆掉可以加上新prefix避免重覆
func getGameSequenceNumber(selfDefine *entity.BaseSelfDefine) string {
	gameSequenceNumberPrefix := mconfig.GetString("api.game.gameSequenceNumberPrefix")
	seqNo := database.GetGameSequenceNumber(selfDefine.TraceID, gameSequenceNumberPrefix)
	//get game sequence number error
	if seqNo == "" {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return ""
	}

	return seqNo
}

func (service *GetSequenceNumberService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
