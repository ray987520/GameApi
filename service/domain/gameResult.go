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

type GameResultService struct {
	Request entity.GameResultRequest
}

// databinding&validate
func ParseGameResultRequest(traceId string, r *http.Request) (request entity.GameResultRequest) {
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

func (service *GameResultService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//parse game token
	account, currency, _ := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	//get wallet
	wallet, isOK := getPlayerWallet(&service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, service.Request.TraceID, innererror.DataNode, tracer.MergeMessage("wallet", wallet))

	//add gameresult
	isAddGameResultOK := addGameResultRecountWallet(&service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddGameResultOK {
		return nil
	}

	//clear wallet cache & get wallet
	refeshData := refreshWallet(&service.Request.BaseSelfDefine, account, currency, service.Request.Token, service.Request.TurnTimes)
	if refeshData == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return refeshData
}

// 寫入RollOut|RollIn|GameResult,更新錢包(transaction)
func addGameResultRecountWallet(selfDefine *entity.BaseSelfDefine, data entity.GameResult, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddGameResultReCountWallet(selfDefine.TraceID, data, wallet, es.LocalNow(8))
	//db access error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}

	return isOK
}

// 更新錢包|betCount緩存
func refreshWallet(selfDefine *entity.BaseSelfDefine, account, currency, token string, betTimes int) interface{} {
	//clear wallet cache
	database.ClearPlayerWalletCache(selfDefine.TraceID, currency, account)

	betCount := database.GetAccountBetCount(selfDefine.TraceID, account)
	//get betcount error
	if betCount == -1 {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}

	wallet, isOK := getPlayerWallet(selfDefine, account, currency)
	//get wallet error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, selfDefine.TraceID, innererror.DataNode, tracer.MergeMessage("wallet", wallet))

	return entity.GameResultResponse{
		Currency: wallet.Currency,
		Amount:   wallet.Amount,
		BetCount: int(betCount),
	}
}

func (service *GameResultService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
