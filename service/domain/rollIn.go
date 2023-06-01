package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

type RollInService struct {
	Request  entity.RollInRequest
	TraceMap string
}

// databinding&validate
func ParseRollInRequest(traceMap string, r *http.Request) (request entity.RollInRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
	}
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) || !strings.HasPrefix(request.TransID, "rollIn-") {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *RollInService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
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
	hasRollOut, rollOutAmount := hasRollOutHistory(es.AddTraceMap(service.TraceMap, string(functionid.HasRollOutHistory)), &service.Request.BaseSelfDefine, service.Request.GameSequenceNumber)
	if !hasRollOut {
		return
	}
	//TOCHECK 理論上rollout amount應等於rollint bet
	if !rollOutAmount.Equal(service.Request.CurrencyKindBet) {
		service.Request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	wallet, isOK := getPlayerWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerWallet)), &service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return
	}
	isAddGameResultOK := addGameResult(es.AddTraceMap(service.TraceMap, string(functionid.AddGameResult)), &service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return
	}
	isAddRollHistoryOK := addRollInHistory(es.AddTraceMap(service.TraceMap, string(functionid.AddRollInHistory)), &service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return
	}
	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), currency, account)
	data = refreshWallet(es.AddTraceMap(service.TraceMap, string(functionid.RefreshWallet)), &service.Request.BaseSelfDefine, account, currency, service.Request.Token, service.Request.TurnTimes)
	return
}

// 防呆,檢查要先有rollOut
func hasRollOutHistory(traceMap string, selfDefine *entity.BaseSelfDefine, seqNo string) (hasData bool, rollOutAmount decimal.Decimal) {
	hasData, rollOutAmount = database.IsExistsRolloutHistory(es.AddTraceMap(traceMap, sqlid.IsExistsRolloutHistory.String()), seqNo)
	if !hasData {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *RollInService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
