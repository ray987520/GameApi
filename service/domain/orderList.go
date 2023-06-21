package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"fmt"
	"net/http"
)

type OrderListService struct {
	Request entity.OrderListRequest
}

// databinding&validate
func ParseOrderListRequest(traceId string, r *http.Request) (request entity.OrderListRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//read query string
	query := r.URL.Query()
	request.Token = query.Get("connectToken")

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *OrderListService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//parse game toekn
	_, _, gameId := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)
	//wrong gameId
	if gameId == 0 {
		return nil
	}

	//get game lang
	lang := getGameLanguage(&service.Request.BaseSelfDefine, gameId)
	//no lang-Code
	if lang == "" {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.OrderListResponse{
		Url: getReportUrl(service.Request.TraceID, lang, service.Request.Token),
	}
}

// 用gameId取出遊戲的語系
func getGameLanguage(selfDefine *entity.BaseSelfDefine, gameId int) string {
	lang := database.GetGameLanguage(selfDefine.TraceID, gameId)
	//get game language error
	if lang == "" {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return ""
	}

	return lang
}

// 組合歷史紀錄網址
func getReportUrl(traceMap string, lang, token string) string {
	reportUrl := mconfig.GetString("api.report.historyReportUrl")

	//如果是zh-CN/zh-TW統一為zh-CN,其他預設en-US
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
