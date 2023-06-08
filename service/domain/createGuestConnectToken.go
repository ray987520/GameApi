package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	es "TestAPI/external/service"
	"fmt"
	"net/http"
	"strconv"
)

type CreateGuestConnectTokenService struct {
	Request  entity.CreateGuestConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseCreateGuestConnectTokenRequest(traceMap string, r *http.Request) (request entity.CreateGuestConnectTokenRequest, err error) {
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
	gameId, err := strconv.Atoi(query.Get("gameID"))
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	request.GameId = gameId

	//validate request
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *CreateGuestConnectTokenService) Exec() interface{} {
	//catch panic
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	//gen a token
	token := genConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.GenConnectToken)), &service.Request.BaseSelfDefine, service.Request.Account, service.Request.Currency, service.Request.GameId)
	if token == "" {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return entity.CreateGuestConnectTokenResponse{
		Token: token,
	}
}

// 創建token,aes128加密,包含gameId,currency,account跟過期時間600秒(timestamp)
func genConnectToken(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) string {
	//new token data,Key:[gameId]_[currency]_[account],ExpitreTime:now+600Sec.
	tokenData := entity.ConnectToken{
		Key:         fmt.Sprintf("%d_%s_%s", gameId, currency, account),
		ExpitreTime: es.Timestamp() + 600,
	}

	//token data to json []byte
	data, err := es.JsonMarshal(es.AddTraceMap(traceMap, string(esid.JsonMarshal)), tokenData)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return ""
	}

	//aes128 encrypt token data,return encrypted string
	token, err := es.Aes128Encrypt(es.AddTraceMap(traceMap, string(esid.Aes128Encrypt)), data)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return ""
	}
	return token
}

func (service *CreateGuestConnectTokenService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
