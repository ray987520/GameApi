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
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	query := r.URL.Query()
	request.FromDate = query.Get("fromDate")
	request.ToDate = query.Get("toDate")
	fromTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, request.FromDate)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	toTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, request.ToDate)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	//開始結束時間不正常或間隔超過24H
	if toTime.Before(fromTime) || toTime.After(fromTime.Add(24*time.Hour)) {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseRoundCheckRequest, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "request.FromDate", request.FromDate, "request.ToDate", request.ToDate)
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *RoundCheckService) Exec() (data interface{}) {
	resp := entity.RoundCheckResponse{
		RoundCheckList: []entity.RoundCheckToken{},
	}
	if service.Request.HasError() {
		return
	}
	roundCheckList, isOK := getRoundCheckList(es.AddTraceMap(service.TraceMap, string(functionid.GetRoundCheckList)), &service.Request.BaseSelfDefine, service.Request.FromDate, service.Request.ToDate)
	if !isOK {
		return
	}
	resp.RoundCheckList = roundCheckList
	data = resp
	return
}

// 取出須補單token並建立僅能補單的cache
func getRoundCheckList(traceMap string, selfDefine *entity.BaseSelfDefine, fromDate, toDate string) (list []entity.RoundCheckToken, isOK bool) {
	list, err := database.GetRoundCheckList(es.AddTraceMap(traceMap, sqlid.GetRoundCheckList.String()), fromDate, toDate)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	//不存在GameResult/RollIn的建立30分鐘token(redis),finishGameResultToken:[token]
	for _, checkData := range list {
		if isOK = database.SetFinishGameResultTokenCache(es.AddTraceMap(traceMap, redisid.SetFinishGameResultTokenCache.String()), checkData.Token); !isOK {
			selfDefine.ErrorCode = string(errorcode.UnknowError)
			return
		}
	}
	isOK = true
	return
}

func (service *RoundCheckService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
