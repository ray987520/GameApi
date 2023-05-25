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
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	query := r.URL.Query()
	request.Account = query.Get("account")
	request.Currency = query.Get("currency")
	gameId, err := strconv.Atoi(query.Get("gameID"))
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	request.GameId = gameId
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *CreateGuestConnectTokenService) Exec() (data interface{}) {
	data = entity.CreateGuestConnectTokenResponse{}
	if service.Request.HasError() {
		return
	}
	token := genConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.GenConnectToken)), &service.Request.BaseSelfDefine, service.Request.Account, service.Request.Currency, service.Request.GameId)
	data = entity.CreateGuestConnectTokenResponse{
		Token: token,
	}
	return
}

// 創建token,aes128加密,包含gameId,currency,account跟過期時間600秒(timestamp)
func genConnectToken(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) (token string) {
	tokenData := entity.ConnectToken{
		Key:         fmt.Sprintf("%d_%s_%s", gameId, currency, account),
		ExpitreTime: es.Timestamp() + 600,
	}
	data, err := es.JsonMarshal(es.AddTraceMap(traceMap, string(esid.JsonMarshal)), tokenData)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	token, err = es.Aes128Encrypt(es.AddTraceMap(traceMap, string(esid.Aes128Encrypt)), data)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *CreateGuestConnectTokenService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
