package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"encoding/json"
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
	json.Unmarshal(body, &request)
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GameResultService) Exec() (data interface{}) {
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
		es.Error("traceMap:%s ,addGameResultAndRollHistory error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

// 更新錢包|betCount緩存
func refreshWallet(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency, token string, betTimes int) interface{} {
	database.ClearPlayerWalletCache(es.AddTraceMap(traceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	betCount := database.IncrConnectTokenBetCount(es.AddTraceMap(traceMap, redisid.IncrConnectTokenBetCount.String()), token, betTimes)
	wallet, isOK := getPlayerWallet(es.AddTraceMap(traceMap, string(functionid.GetPlayerWallet)), selfDefine, account, currency)
	if !isOK {
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
