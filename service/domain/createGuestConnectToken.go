package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	es "TestAPI/external/service"
	"TestAPI/external/service/str"
	"TestAPI/external/service/tracer"
	"fmt"
	"net/http"
)

type CreateGuestConnectTokenService struct {
	Request entity.CreateGuestConnectTokenRequest
}

// databinding&validate
func ParseCreateGuestConnectTokenRequest(traceId string, r *http.Request) (request entity.CreateGuestConnectTokenRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//read query
	query := r.URL.Query()
	request.Account = query.Get("account")
	request.Currency = query.Get("currency")
	gameId, isOK := str.Atoi(traceId, query.Get("gameID"))
	//convert error
	if !isOK {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	request.GameId = gameId

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *CreateGuestConnectTokenService) Exec() interface{} {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//gen a token
	token := genConnectToken(&service.Request.BaseSelfDefine, service.Request.Account, service.Request.Currency, service.Request.GameId)
	if token == "" {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.CreateGuestConnectTokenResponse{
		Token: token,
	}
}

// 創建token,aes128加密,包含gameId,currency,account跟過期時間600秒(timestamp)
func genConnectToken(selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) string {
	//new token data,Key:[gameId]_[currency]_[account],ExpitreTime:now+600Sec.
	tokenData := entity.ConnectToken{
		Key:         fmt.Sprintf("%d_%s_%s", gameId, currency, account),
		ExpitreTime: es.Timestamp() + 600,
	}

	//token data to json []byte
	data := es.JsonMarshal(selfDefine.TraceID, tokenData)
	if data == nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return ""
	}

	//aes128 encrypt token data,return encrypted string
	token := es.Aes128Encrypt(selfDefine.TraceID, data)
	if token == "" {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return ""
	}
	return token
}

func (service *CreateGuestConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
