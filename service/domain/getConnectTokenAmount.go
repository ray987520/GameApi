package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	es "TestAPI/external/service"
	"net/http"
)

type GetConnectTokenAmountService struct {
	Request  entity.GetConnectTokenAmountRequest
	TraceMap string
}

// databinding&validate
func ParseGetConnectTokenAmountRequest(traceMap string, r *http.Request) (request entity.GetConnectTokenAmountRequest, err error) {
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	query := r.URL.Query()
	request.Token = query.Get("connectToken")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GetConnectTokenAmountService) Exec() (data interface{}) {
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
	//don't show WalletID
	wallet.WalletID = ""
	data = wallet
	return
}

func (service *GetConnectTokenAmountService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
