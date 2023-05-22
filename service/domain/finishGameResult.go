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

type FinishGameResultService struct {
	Request  entity.FinishGameResultRequest
	TraceMap string
}

// databinding&validate
func ParseFinishGameResultRequest(traceMap string, r *http.Request) (request entity.FinishGameResultRequest, err error) {
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

func (service *FinishGameResultService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	if isFinishGameResultConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsFinishGameResultConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	account, currency, _ := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return
	}
	//取用戶錢包,id會用來做rowlock
	wallet, isOK := getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return
	}
	//不存在GameResult的話寫GameResult跟RollIn,不存在RollIn只寫RollIn
	isAddGameResultOK := addGameResult(es.AddTraceMap(service.TraceMap, string(functionid.AddGameResult)), &service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return
	}
	isAddRollHistoryOK := addRollInHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddRollInHistory)), &service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return
	}
	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), currency, account)

	return
}

// 檢查補單aes token活耀
func isFinishGameResultConnectTokenAlive(traceMap string, token string) bool {
	alive := database.GetFinishGameResultTokenAlive(es.AddTraceMap(traceMap, sqlid.GetFinishGameResultTokenAlive.String()), token)
	if !alive {
		es.Error("traceMap:%s ,token is dead ,token:%s", traceMap, token)
	}
	return alive
}

// 判斷補單connectToken是否正常
func isFinishGameResultConnectTokenError(traceMap string, selfDefine *entity.BaseSelfDefine, token string) bool {
	if !isFinishGameResultConnectTokenAlive(es.AddTraceMap(traceMap, string(functionid.IsFinishGameResultConnectTokenAlive)), token) {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return true
	}
	return false
}

// 取玩家錢包
func getPlayerWallet(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string) (data entity.PlayerWallet, isOK bool) {
	data, err := database.GetPlayerWallet(es.AddTraceMap(traceMap, sqlid.GetPlayerWallet.String()), account, currency)
	if err != nil || data.Currency == "" {
		es.Error("traceMap:%s ,error:%v ,data.Currency:%s", traceMap, err, data.Currency)
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	isOK = true
	return
}

// 補寫GameResult
func addGameResult(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult) (isOK bool) {
	hasGameResult := database.IsExistsTokenGameResult(es.AddTraceMap(traceMap, sqlid.IsExistsTokenGameResult.String()), data.Token, data.GameSequenceNumber)
	if hasGameResult {
		isOK = true
		return
	}
	isOK = database.AddGameResult(es.AddTraceMap(traceMap, sqlid.AddGameResult.String()), data)
	if !isOK {
		es.Error("traceMap:%s ,addGameResult error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

// 補寫RollIn
func addRollInHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	hasRollInHistory := database.IsExistsRollInHistory(es.AddTraceMap(traceMap, sqlid.IsExistsRollInHistory.String()), data.Token, data.GameSequenceNumber)
	if hasRollInHistory {
		isOK = true
		return
	}
	isOK = database.AddRollInHistory(es.AddTraceMap(traceMap, sqlid.AddRollInHistory.String()), data, wallet, es.LocalNow(8))
	if !isOK {
		es.Error("traceMap:%s ,addRollInHistory error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *FinishGameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
