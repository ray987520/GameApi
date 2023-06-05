package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"net/http"
)

type FinishGameResultService struct {
	Request  entity.FinishGameResultRequest
	TraceMap string
}

// databinding&validate
func ParseFinishGameResultRequest(traceMap string, r *http.Request) (request entity.FinishGameResultRequest, err error) {
	body, err := readHttpRequestBody(es.AddTraceMap(traceMap, string(functionid.ReadHttpRequestBody)), r, &request)
	if err != nil {
		return request, err
	}

	err = parseJsonBody(es.AddTraceMap(traceMap, string(functionid.ParseJsonBody)), body, &request)
	if err != nil {
		return request, err
	}

	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *FinishGameResultService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	if isFinishGameResultConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsFinishGameResultConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	account, currency, _ := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	//取用戶錢包,id會用來做rowlock
	wallet, isOK := getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	//不存在GameResult的話寫GameResult跟RollIn,不存在RollIn只寫RollIn
	isAddGameResultOK := addGameResult(es.AddTraceMap(service.TraceMap, string(functionid.AddGameResult)), &service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return nil
	}

	isAddRollHistoryOK := addRollInHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddRollInHistory)), &service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	return
}

// 檢查補單aes token活耀
func isFinishGameResultConnectTokenAlive(traceMap string, token string) bool {
	alive := database.GetFinishGameResultTokenAlive(es.AddTraceMap(traceMap, sqlid.GetFinishGameResultTokenAlive.String()), token)
	return alive
}

// 判斷補單connectToken是否正常
func isFinishGameResultConnectTokenError(traceMap string, selfDefine *entity.BaseSelfDefine, token string) (alive bool) {
	alive = isFinishGameResultConnectTokenAlive(es.AddTraceMap(traceMap, string(functionid.IsFinishGameResultConnectTokenAlive)), token)
	if !alive {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
	}
	return alive
}

// 取玩家錢包
func getPlayerWallet(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string) (data entity.PlayerWallet, isOK bool) {
	data, err := database.GetPlayerWallet(es.AddTraceMap(traceMap, sqlid.GetPlayerWallet.String()), account, currency)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return entity.PlayerWallet{}, false
	}
	return data, true
}

// 補寫GameResult
func addGameResult(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult) (isOK bool) {
	hasGameResult := database.IsExistsTokenGameResult(es.AddTraceMap(traceMap, sqlid.IsExistsTokenGameResult.String()), data.Token, data.GameSequenceNumber)
	if hasGameResult {
		return true
	}

	isOK = database.AddGameResult(es.AddTraceMap(traceMap, sqlid.AddGameResult.String()), data)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 補寫RollIn
func addRollInHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	hasRollInHistory := database.IsExistsRollInHistory(es.AddTraceMap(traceMap, sqlid.IsExistsRollInHistory.String()), data.Token, data.GameSequenceNumber)
	if hasRollInHistory {
		return true
	}

	isOK = database.AddRollInHistory(es.AddTraceMap(traceMap, sqlid.AddRollInHistory.String()), data, wallet, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *FinishGameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
