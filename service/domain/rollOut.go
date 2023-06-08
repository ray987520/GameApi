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
	"strings"
)

type RollOutService struct {
	Request  entity.RollOutRequest
	TraceMap string
}

// databinding&validate
func ParseRollOutRequest(traceMap string, r *http.Request) (request entity.RollOutRequest, err error) {
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

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) || !strings.HasPrefix(request.TransID, "rollOut-") {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *RollOutService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)

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

	isAddRollHistoryOK := addRollOutHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddRollOutHistory)), &service.Request.BaseSelfDefine, service.Request.RollHistory, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), currency, account)

	if service.Request.TakeAll == 0 {
		wallet, isOK = getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
		if !isOK {
			return nil
		}

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
func addRollOutHistory(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.RollHistory, wallet entity.PlayerWallet) (isOK bool) {
	isOK = database.AddRollOutHistory(es.AddTraceMap(traceMap, sqlid.AddRollOutHistory.String()), data, wallet)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *RollOutService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
