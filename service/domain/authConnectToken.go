package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	"fmt"
	"io/ioutil"
	"net/http"
)

type AuthConnectTokenService struct {
	Request  entity.AuthConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseAuthConnectTokenRequest(traceMap string, r *http.Request) (request entity.AuthConnectTokenRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
		return
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *AuthConnectTokenService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	account, currency, gameId := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, false)
	if account == "" {
		return
	}
	//token沒問題的話,撈playerInfo,正確的話寫進GameToken,並放進redis
	playerInfo, isOK := getPlayerInfo(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerInfo)), &service.Request.BaseSelfDefine, account, currency, gameId)
	if !isOK {
		return
	}
	if isOK := addConnectToken2Db(es.AddTraceMap(service.TraceMap, string(functionid.AddConnectToken2Db)), &service.Request.BaseSelfDefine, service.Request.Token, account, currency, service.Request.Ip, gameId); !isOK {
		return
	}
	//PlayerWallet的WalletID/Currency後續有使用但不輸出於此response
	playerInfo.WalletID = ""
	playerInfo.PlayerWallet.Currency = ""
	data = playerInfo
	return
}

// 解密aes token,若超過expiretime能解密也無效
func parseConnectToken(traceMap string, selfDefine *entity.BaseSelfDefine, token string, passExpire bool) (account, currency string, gameId int) {
	var tokenData entity.ConnectToken
	data, err := es.Aes128Decrypt(es.AddTraceMap(traceMap, string(esid.Aes128Decrypt)), token)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), data, &tokenData)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), tokenData) {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	if !passExpire {
		now := es.Timestamp()
		if now > tokenData.ExpitreTime {
			err = fmt.Errorf("token expired")
			zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseConnectToken, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
			selfDefine.ErrorCode = string(errorcode.BadParameter)
			return
		}
	}
	account, currency, gameId = tokenData.Parse()
	return account, currency, gameId
}

// ConnectToken添加到GameToken
func addConnectToken2Db(traceMap string, selfDefine *entity.BaseSelfDefine, token, account, currency, ip string, gameId int) (isOK bool) {
	isOK = database.AddConnectToken(es.AddTraceMap(traceMap, sqlid.AddConnectToken.String()), token, account, currency, ip, gameId, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

// 取PlayerInfo(Base|BetCount|Wallet)
func getPlayerInfo(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) (playerInfo entity.AuthConnectTokenResponse, isOK bool) {
	playerInfo, err := database.GetPlayerInfo(es.AddTraceMap(traceMap, sqlid.GetPlayerInfo.String()), account, currency, gameId)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		isOK = false
		return
	}
	isOK = true
	return
}

func (service *AuthConnectTokenService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
