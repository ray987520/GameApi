package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"net/http"
)

type GetConnectTokenAmountService struct {
	Request entity.GetConnectTokenAmountRequest
}

// databinding&validate
func ParseGetConnectTokenAmountRequest(traceId string, r *http.Request) (request entity.GetConnectTokenAmountRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//read querystring
	query := r.URL.Query()
	request.Token = query.Get("connectToken")

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *GetConnectTokenAmountService) Exec() (data interface{}) {
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

	//don't show WalletID
	wallet.WalletID = ""
	service.Request.ErrorCode = string(errorcode.Success)
	return wallet
}

func (service *GetConnectTokenAmountService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
