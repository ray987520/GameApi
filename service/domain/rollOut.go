package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"net/http"
	"strings"
)

type RollOutService struct {
	Request entity.RollOutRequest
}

// databinding&validate
func ParseRollOutRequest(traceId string, r *http.Request) (request entity.RollOutRequest) {
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

	//validate reqeust,transId因為model公用所以在這裡另外寫條件
	if !IsValid(traceId, request) || !strings.HasPrefix(request.TransID, "rollOut-") {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	return request
}

func (service *RollOutService) Exec() (data interface{}) {
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

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, service.Request.TraceID, "wallet", wallet)

	//add rollOut record
	isAddRollHistoryOK := addRollOutHistory(&service.Request.BaseSelfDefine, service.Request.RollHistory, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	//clear wallet cache to sync
	database.ClearPlayerWalletCache(service.Request.TraceID, currency, account)

	//if user want get wallet data
	if service.Request.TakeAll == 0 {
		wallet, isOK = getPlayerWallet(&service.Request.BaseSelfDefine, account, currency)
		//get wallet error
		if !isOK {
			return nil
		}

		zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, service.Request.TraceID, "wallet", wallet)

		service.Request.ErrorCode = string(errorcode.Success)
		return entity.RollOutResponse{
			Currency: wallet.Currency,
			Amount:   service.Request.Amount,
			Balance:  wallet.Amount,
		}
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 添加rollOut並更新錢包
func addRollOutHistory(selfDefine *entity.BaseSelfDefine, data entity.RollHistory, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddRollOutHistory(selfDefine.TraceID, data, wallet)
	//add rollOut record error
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *RollOutService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
