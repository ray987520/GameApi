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
)

type GameResultService struct {
	Request  entity.GameResultRequest
	TraceMap string
}

// databinding&validate
func ParseGameResultRequest(traceMap string, r *http.Request) (request entity.GameResultRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
	}
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GameResultService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
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
	isAddGameResultOK := addGameResultAndRollHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddGameResultAndRollHistory)), &service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddGameResultOK {
		return
	}
	data = refreshWallet(es.AddTraceMap(service.TraceMap, string(functionid.RefreshWallet)), &service.Request.BaseSelfDefine, account, currency, service.Request.Token, service.Request.TurnTimes)
	return
}

// 寫入RollOut|RollIn|GameResult,更新錢包(transaction)
func addGameResultAndRollHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddGameResultReCountWallet(es.AddTraceMap(traceMap, sqlid.AddGameResultReCountWallet.String()), data, wallet, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

// 更新錢包|betCount緩存
func refreshWallet(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency, token string, betTimes int) interface{} {
	database.ClearPlayerWalletCache(es.AddTraceMap(traceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	betCount, err := database.GetAccountBetCount(es.AddTraceMap(traceMap, sqlid.GetAccountBetCount.String()), token)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	wallet, isOK := getPlayerWallet(es.AddTraceMap(traceMap, string(functionid.GetPlayerWallet)), selfDefine, account, currency)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	return entity.GameResultResponse{
		Currency: wallet.Currency,
		Amount:   wallet.Amount,
		BetCount: int(betCount),
	}
}

func (service *GameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
