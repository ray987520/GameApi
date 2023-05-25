package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"io/ioutil"
	"net/http"
	"strings"
)

type RollOutService struct {
	Request  entity.RollOutRequest
	TraceMap string
}

// databinding&validate
func ParseRollOutRequest(traceMap string, r *http.Request) (request entity.RollOutRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
	}
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) || !strings.HasPrefix(request.TransID, "rollOut-") {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *RollOutService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	account, currency, _ := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return
	}
	wallet, isOK := getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return
	}
	isAddRollHistoryOK := addRollOutHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddRollOutHistory)), &service.Request.BaseSelfDefine, service.Request.RollHistory, wallet)
	if !isAddRollHistoryOK {
		return
	}
	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	if service.Request.TakeAll == 0 {
		wallet, isOK = getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
		if !isOK {
			return
		}
		data = entity.RollOutResponse{
			Currency: wallet.Currency,
			Amount:   service.Request.Amount,
			Balance:  wallet.Amount,
		}
	}
	return
}

// 添加rollOut並更新錢包
func addRollOutHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.RollHistory, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddRollOutHistory(es.AddTraceMap(traceMap, sqlid.AddRollOutHistory.String()), data, wallet)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return
}

func (service *RollOutService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
