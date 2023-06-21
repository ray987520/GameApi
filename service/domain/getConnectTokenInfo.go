package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"net/http"
)

type GetConnectTokenInfoService struct {
	Request entity.GetConnectTokenInfoRequest
}

// databinding&validate
func ParseGetConnectTokenInfoRequest(traceId string, r *http.Request) (request entity.GetConnectTokenInfoRequest) {
	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//read query string
	query := r.URL.Query()
	request.Token = query.Get("connectToken")

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *GetConnectTokenInfoService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//parse game token
	account, currency, gameId := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)

	//get playerinfo
	data = getPlayerInfoCache(&service.Request.BaseSelfDefine, account, currency, gameId)
	if data == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return data
}

// 取出緩存PlayerInfo
func getPlayerInfoCache(selfDefine *entity.BaseSelfDefine, account, currency string, gameId int) interface{} {
	playerInfo := database.GetPlayerInfo(selfDefine.TraceID, account, currency, gameId)
	//get playerinfo error
	if playerInfo.GameAccount == "" {
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
