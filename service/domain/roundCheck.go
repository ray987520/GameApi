package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	"net/http"
	"time"
)

type RoundCheckService struct {
	Request  entity.RoundCheckRequest
	TraceMap string
}

// databinding&validate
func ParseRoundCheckRequest(traceMap string, r *http.Request) (request entity.RoundCheckRequest, err error) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	query := r.URL.Query()
	request.FromDate = query.Get("fromDate")
	request.ToDate = query.Get("toDate")
	fromTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, request.FromDate)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	toTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, request.ToDate)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	//開始結束時間不正常或間隔超過24H
	if toTime.Before(fromTime) || toTime.After(fromTime.Add(24*time.Hour)) {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseRoundCheckRequest, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "request.FromDate", request.FromDate, "request.ToDate", request.ToDate)
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *RoundCheckService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	roundCheckList, isOK := getRoundCheckList(es.AddTraceMap(service.TraceMap, string(functionid.GetRoundCheckList)), &service.Request.BaseSelfDefine, service.Request.FromDate, service.Request.ToDate)
	if !isOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.RoundCheckResponse{
		RoundCheckList: roundCheckList,
	}
}

// 取出須補單token並建立僅能補單的cache
func getRoundCheckList(traceMap string, selfDefine *entity.BaseSelfDefine, fromDate, toDate string) (list []entity.RoundCheckToken, isOK bool) {
	list, err := database.GetRoundCheckList(es.AddTraceMap(traceMap, sqlid.GetRoundCheckList.String()), fromDate, toDate)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil, false
	}

	//不存在GameResult/RollIn的建立30分鐘token(redis),finishGameResultToken:[token]
	for _, checkData := range list {
		isOK = database.SetFinishGameResultTokenCache(es.AddTraceMap(traceMap, redisid.SetFinishGameResultTokenCache.String()), checkData.Token)
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
