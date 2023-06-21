package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"net/http"
	"time"
)

type RoundCheckService struct {
	Request entity.RoundCheckRequest
}

const (
	timeRangeError = "bad time range"
)

// databinding&validate
func ParseRoundCheckRequest(traceId string, r *http.Request) (request entity.RoundCheckRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//read query string
	query := r.URL.Query()
	request.FromDate = query.Get("fromDate")
	request.ToDate = query.Get("toDate")

	fromTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, request.FromDate)
	//parse time error
	if !isOK {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	toTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, request.ToDate)
	//parse time error
	if !isOK {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	//開始結束時間不正常或間隔超過24H,validate套件可以做類似驗證但是拉出來比較清楚
	if toTime.Before(fromTime) || toTime.After(fromTime.Add(24*time.Hour)) {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseRoundCheckRequest, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, timeRangeError, "fromTime", fromTime, "toTime", toTime))
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	return request
}

func (service *RoundCheckService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//取須補單清單
	roundCheckList, isOK := getRoundCheckList(&service.Request.BaseSelfDefine, service.Request.FromDate, service.Request.ToDate)
	if !isOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.RoundCheckResponse{
		RoundCheckList: roundCheckList,
	}
}

// 取出須補單token並建立僅能補單的cache
func getRoundCheckList(selfDefine *entity.BaseSelfDefine, fromDate, toDate string) (list []entity.RoundCheckToken, isOK bool) {
	list = database.GetRoundCheckList(selfDefine.TraceID, fromDate, toDate)
	//db access error
	if list == nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil, false
	}

	//不存在GameResult/RollIn的建立30分鐘token(redis),finishGameResultToken:[token]
	for _, checkData := range list {
		isOK = database.SetFinishGameResultTokenCache(selfDefine.TraceID, checkData.Token)
		//設置補單token cache失敗
		if !isOK {
			selfDefine.ErrorCode = string(errorcode.UnknowError)
			return nil, false
		}
	}

	return list, true
}

func (service *RoundCheckService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
