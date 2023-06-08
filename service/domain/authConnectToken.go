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
	"net/http"
)

type AuthConnectTokenService struct {
	Request  entity.AuthConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseAuthConnectTokenRequest(traceMap string, r *http.Request) (request entity.AuthConnectTokenRequest, err error) {
	body, err := readHttpRequestBody(es.AddTraceMap(traceMap, string(functionid.ReadHttpRequestBody)), r, &request)
	if err != nil {
		return request, err
	}

	err = parseJsonBody(es.AddTraceMap(traceMap, string(functionid.ParseJsonBody)), body, &request)
	if err != nil {
		return request, err
	}

	//read request header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//validate request
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}

	return request, nil
}

func (service *AuthConnectTokenService) Exec() interface{} {
	//catch panic
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	//get account, currency, gameId from token
	account, currency, gameId := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, false)
	if account == "" {
		return nil
	}

	//token沒問題的話,撈playerInfo,正確的話寫進GameToken,並放進redis
	playerInfo, isOK := getPlayerInfo(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerInfo)), &service.Request.BaseSelfDefine, account, currency, gameId)
	if !isOK {
		return nil
	}

	//insert token
	isOK = addConnectToken2Db(es.AddTraceMap(service.TraceMap, string(functionid.AddConnectToken2Db)), &service.Request.BaseSelfDefine, service.Request.Token, account, currency, service.Request.Ip, gameId)
	if !isOK {
		return nil
	}

	//PlayerWallet的WalletID/Currency後續有使用但不輸出於此response
	playerInfo.WalletID = ""
	playerInfo.PlayerWallet.Currency = ""
	service.Request.ErrorCode = string(errorcode.Success)
	return playerInfo
}

// 解密aes token,若超過expiretime能解密也無效
func parseConnectToken(traceMap string, selfDefine *entity.BaseSelfDefine, token string, passExpire bool) (account, currency string, gameId int) {
	var tokenData entity.ConnectToken

	//aes128 decrypt token
	data, err := es.Aes128Decrypt(es.AddTraceMap(traceMap, string(esid.Aes128Decrypt)), token)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//parse decrypted data
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), data, &tokenData)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//validate token model
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), tokenData) {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//pass token expiretime check or not
	if !passExpire {
		now := es.Timestamp()
		if now > tokenData.ExpitreTime {
			err = fmt.Errorf("token expired")
			zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseConnectToken, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
			selfDefine.ErrorCode = string(errorcode.BadParameter)
			return "", "", 0
		}
	}

	account, currency, gameId = tokenData.Parse()
	return account, currency, gameId
}

// ConnectToken添加到GameToken
func addConnectToken2Db(traceMap string, selfDefine *entity.BaseSelfDefine, token, account, currency, ip string, gameId int) bool {
	isOK := database.AddConnectToken(es.AddTraceMap(traceMap, sqlid.AddConnectToken.String()), token, account, currency, ip, gameId, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 取PlayerInfo(Base|BetCount|Wallet)
func getPlayerInfo(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) (entity.AuthConnectTokenResponse, bool) {
	playerInfo, err := database.GetPlayerInfo(es.AddTraceMap(traceMap, sqlid.GetPlayerInfo.String()), account, currency, gameId)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return entity.AuthConnectTokenResponse{}, false
	}
	return playerInfo, true
}

func (service *AuthConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
