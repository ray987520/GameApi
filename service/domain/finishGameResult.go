package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"net/http"
)

type FinishGameResultService struct {
	Request entity.FinishGameResultRequest
}

// databinding&validate
func ParseFinishGameResultRequest(traceId string, r *http.Request) (request entity.FinishGameResultRequest) {
	body, isOK := readHttpRequestBody(traceId, r, &request)
	//read body error
	if !isOK {
		return request
	}

	isOK = parseJsonBody(traceId, body, &request)
	//json deserialize error
	if !isOK {
		return request
	}

	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *FinishGameResultService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isFinishGameResultConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	account, currency, _ := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	//取用戶錢包,id會用來做rowlock
	wallet, isOK := getPlayerWallet(&service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	//不存在GameResult的話寫GameResult跟RollIn,不存在RollIn只寫RollIn
	isAddGameResultOK := addGameResult(&service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return nil
	}

	isAddRollHistoryOK := addRollInHistory(&service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	database.ClearPlayerWalletCache(service.Request.TraceID, currency, account)
	service.Request.ErrorCode = string(errorcode.Success)
	return
}

// 檢查補單aes token活耀
func isFinishGameResultConnectTokenAlive(traceId string, token string) bool {
	alive := database.GetFinishGameResultTokenAlive(traceId, token)
	return alive
}

// 判斷補單connectToken是否正常
func isFinishGameResultConnectTokenError(selfDefine *entity.BaseSelfDefine, token string) (alive bool) {
	alive = isFinishGameResultConnectTokenAlive(selfDefine.TraceID, token)
	if !alive {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
	}
	return alive
}

// 取玩家錢包
func getPlayerWallet(selfDefine *entity.BaseSelfDefine, account, currency string) (data entity.PlayerWallet, isOK bool) {
	data, isOK = database.GetPlayerWallet(selfDefine.TraceID, account, currency)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return entity.PlayerWallet{}, false
	}
	return data, true
}

// 補寫GameResult
func addGameResult(selfDefine *entity.BaseSelfDefine, data entity.GameResult) (isOK bool) {
	hasGameResult := database.IsExistsTokenGameResult(selfDefine.TraceID, data.Token, data.GameSequenceNumber)
	//if game result existed,pass
	if hasGameResult {
		return true
	}

	isOK = database.AddGameResult(selfDefine.TraceID, data)
	//db error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 補寫RollIn
func addRollInHistory(selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	hasRollInHistory := database.IsExistsRollInHistory(selfDefine.TraceID, data.Token, data.GameSequenceNumber)
	//if rollIn history existed,pass
	if hasRollInHistory {
		return true
	}

	isOK = database.AddRollInHistory(selfDefine.TraceID, data, wallet, es.LocalNow(8))
	//db error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *FinishGameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
