package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
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
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

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

	//檢查補單token存活
	if isFinishGameResultConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//解析game token
	account, currency, _ := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	//取用戶錢包,id會用來做rowlock
	wallet, isOK := getPlayerWallet(&service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, service.Request.TraceID, innererror.DataNode, tracer.MergeMessage("wallet", wallet))

	//不存在GameResult的話寫GameResult跟RollIn,不存在RollIn只寫RollIn
	isAddGameResultOK := addGameResult(&service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return nil
	}

	//add rollIn記錄
	isAddRollHistoryOK := addRollInHistory(&service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	//清除錢包cache
	database.ClearPlayerWalletCache(service.Request.TraceID, currency, account)

	service.Request.ErrorCode = string(errorcode.Success)
	return
}

// 檢查補單aes token活耀
func isFinishGameResultConnectTokenAlive(traceId string, token string) bool {
	return database.GetFinishGameResultTokenAlive(traceId, token)
}

// 判斷補單connectToken是否正常
func isFinishGameResultConnectTokenError(selfDefine *entity.BaseSelfDefine, token string) (alive bool) {
	alive = isFinishGameResultConnectTokenAlive(selfDefine.TraceID, token)
	//補單token異常
	if !alive {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
	}

	return alive
}

// 取玩家錢包
func getPlayerWallet(selfDefine *entity.BaseSelfDefine, account, currency string) (data entity.PlayerWallet, isOK bool) {
	data, isOK = database.GetPlayerWallet(selfDefine.TraceID, account, currency)
	//get wallet error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return entity.PlayerWallet{}, false
	}

	return data, true
}

// 補寫GameResult
func addGameResult(selfDefine *entity.BaseSelfDefine, data entity.GameResult) (isOK bool) {
	hasGameResult := database.IsExistsTokenGameResult(selfDefine.TraceID, data.Token, data.GameSequenceNumber)
	//if game result existed,pass this action
	if hasGameResult {
		return true
	}

	isOK = database.AddGameResult(selfDefine.TraceID, data)
	//add game result error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}

	return isOK
}

// 補寫RollIn
func addRollInHistory(selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	hasRollInHistory := database.IsExistsRollInHistory(selfDefine.TraceID, data.Token, data.GameSequenceNumber)
	//if rollIn history existed,pass this action
	if hasRollInHistory {
		return true
	}

	isOK = database.AddRollInHistory(selfDefine.TraceID, data, wallet, es.LocalNow(8))
	//add rollIn record error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *FinishGameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
