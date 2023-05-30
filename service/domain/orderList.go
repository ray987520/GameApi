package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"fmt"
	"net/http"
)

type OrderListService struct {
	Request  entity.OrderListRequest
	TraceMap string
}

// databinding&validate
func ParseOrderListRequest(traceMap string, r *http.Request) (request entity.OrderListRequest, err error) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	query := r.URL.Query()
	request.Token = query.Get("connectToken")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *OrderListService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
	if service.Request.HasError() {
		return
	}
	_, _, gameId := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	if gameId == 0 {
		return
	}
	lang := getGameLanguage(es.AddTraceMap(service.TraceMap, string(functionid.GetGameLanguage)), &service.Request.BaseSelfDefine, gameId)
	if lang == "" {
		return
	}
	data = entity.OrderListResponse{
		Url: getReportUrl(es.AddTraceMap(service.TraceMap, string(functionid.GetReportUrl)), lang, service.Request.Token),
	}
	return
}

// 用gameId取出遊戲的語系
func getGameLanguage(traceMap string, selfDefine *entity.BaseSelfDefine, gameId int) string {
	lang, err := database.GetGameLanguage(es.AddTraceMap(traceMap, sqlid.GetGameLanguage.String()), gameId)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return ""
	}
	return lang
}

// 組合歷史紀錄網址
func getReportUrl(traceMap string, lang, token string) string {
	reportUrl := "https://www.gameReport.com"
	if lang == "zh-CN" || lang == "zh-TW" {
		lang = "zh-CN"
	} else {
		lang = "en-US"
	}
	return fmt.Sprintf("%s?gametoken=%s&language=%s", reportUrl, token, lang)
}

func (service *OrderListService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
