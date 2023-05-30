package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"net/http"
)

type GetConnectTokenInfoService struct {
	Request  entity.GetConnectTokenInfoRequest
	TraceMap string
}

// databinding&validate
func ParseGetConnectTokenInfoRequest(traceMap string, r *http.Request) (request entity.GetConnectTokenInfoRequest, err error) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	query := r.URL.Query()
	request.Token = query.Get("connectToken")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GetConnectTokenInfoService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	account, currency, gameId := parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true)
	data = getPlayerInfoCache(es.AddTraceMap(service.TraceMap, string(functionid.GetPlayerInfoCache)), &service.Request.BaseSelfDefine, account, currency, gameId)
	return
}

// 取出緩存PlayerInfo
func getPlayerInfoCache(traceMap string, selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) interface{} {
	playerInfo, err := database.GetPlayerInfo(es.AddTraceMap(traceMap, sqlid.GetPlayerInfo.String()), account, currency, gameId)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}
	base := playerInfo.PlayerBase
	wallet := playerInfo.PlayerWallet
	//不輸出walletId跟重複的Currency欄位
	wallet.WalletID = ""
	wallet.Currency = ""
	return entity.GetConnectTokenInfoResponse{
		PlatformID:   base.PlatformID,
		ChannelID:    base.ChannelID,
		MemberID:     base.MemberID,
		GameID:       base.GameID,
		GameAccount:  base.GameAccount,
		Currency:     base.Currency,
		PlayerWallet: wallet,
	}
}

func (service *GetConnectTokenInfoService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
