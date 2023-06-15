package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"net/http"
)

type GameResultService struct {
	Request  entity.GameResultRequest
	TraceMap string
}

// databinding&validate
func ParseGameResultRequest(traceMap string, r *http.Request) (request entity.GameResultRequest, err error) {
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

func (service *GameResultService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	account, currency, _ := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	wallet, isOK := getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	isAddGameResultOK := addGameResultAndRollHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddGameResultAndRollHistory)), &service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddGameResultOK {
		return nil
	}

	refeshData := refreshWallet(es.AddTraceMap(service.TraceMap, string(functionid.RefreshWallet)), &service.Request.BaseSelfDefine, account, currency, service.Request.Token, service.Request.TurnTimes)
	if refeshData == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return refeshData
}

// 寫入RollOut|RollIn|GameResult,更新錢包(transaction)
func addGameResultAndRollHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddGameResultReCountWallet(es.AddTraceMap(traceMap, sqlid.AddGameResultReCountWallet.String()), data, wallet, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 更新錢包|betCount緩存
func refreshWallet(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency, token string, betTimes int) interface{} {
	database.ClearPlayerWalletCache(es.AddTraceMap(traceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	betCount, err := database.GetAccountBetCount(es.AddTraceMap(traceMap, sqlid.GetAccountBetCount.String()), account)
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
