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
	"fmt"
	"net/http"
)

type AuthConnectTokenService struct {
	Request entity.AuthConnectTokenRequest
}

const (
	tokenExpire = "token expired"
)

// databinding&validate
func ParseAuthConnectTokenRequest(traceId string, r *http.Request) (request entity.AuthConnectTokenRequest) {
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

	//read request header
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

func (service *AuthConnectTokenService) Exec() interface{} {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//get account, currency, gameId from token
	account, currency, gameId := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, false)
	if account == "" {
		return nil
	}

	//token沒問題的話,撈playerInfo,正確的話寫進GameToken,並放進redis
	playerInfo, isOK := getPlayerInfo(&service.Request.BaseSelfDefine, account, currency, gameId)
	if !isOK {
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerInfo, innererror.TraceNode, service.Request.TraceID, innererror.DataNode, tracer.MergeMessage("playerInfo", playerInfo))

	//insert token
	isOK = addConnectToken2Db(&service.Request.BaseSelfDefine, service.Request.Token, account, currency, service.Request.Ip, gameId)
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
func parseConnectToken(selfDefine *entity.BaseSelfDefine, token string, passExpire bool) (account, currency string, gameId int) {
	var tokenData entity.ConnectToken

	//aes128 decrypt token
	data := es.Aes128Decrypt(selfDefine.TraceID, token)
	if data == nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//parse decrypted data
	isOK := es.JsonUnMarshal(selfDefine.TraceID, data, &tokenData)
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//validate token model
	if !IsValid(selfDefine.TraceID, tokenData) {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", "", 0
	}

	//pass token expiretime check or not
	if !passExpire {
		now := es.Timestamp()
		if now > tokenData.ExpitreTime {
			err := fmt.Errorf(tokenExpire)
			zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ParseConnectToken, innererror.TraceNode, selfDefine.TraceID, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, err, "now", now, "tokenData.ExpitreTime", tokenData.ExpitreTime))
			selfDefine.ErrorCode = string(errorcode.BadParameter)
			return "", "", 0
		}
	}

	//parse token
	account, currency, gameId = tokenData.Parse()

	return account, currency, gameId
}

// ConnectToken添加到GameToken
func addConnectToken2Db(selfDefine *entity.BaseSelfDefine, token, account, currency, ip string, gameId int) bool {
	isOK := database.AddConnectToken(selfDefine.TraceID, token, account, currency, ip, gameId, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

// 取PlayerInfo(Base|BetCount|Wallet)
func getPlayerInfo(selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) (entity.AuthConnectTokenResponse, bool) {
	playerInfo := database.GetPlayerInfo(selfDefine.TraceID, account, currency, gameId)
	if playerInfo.GameAccount == "" {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return entity.AuthConnectTokenResponse{}, false
	}
	return playerInfo, true
}

func (service *AuthConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
