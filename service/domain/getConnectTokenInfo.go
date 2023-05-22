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
)

type GetConnectTokenInfoService struct {
	Request  entity.GetConnectTokenInfoRequest
	TraceMap string
}

// databinding&validate
func ParseGetConnectTokenInfoRequest(traceMap string, r *http.Request) (request entity.GetConnectTokenInfoRequest, err error) {
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	query := r.URL.Query()
	request.Token = query.Get("connectToken")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *GetConnectTokenInfoService) Exec() (data interface{}) {
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
	base, wallet, err := database.GetPlayerInfoCache(es.AddTraceMap(traceMap, redisid.GetPlayerInfoCache.String()), account, currency, gameId)
	//cache壞掉撈db
	if err != nil {
		es.Error("traceMap:%s ,get cache error:%v ,base.MemberAccount:%s ,wallet.Currency:%s", traceMap, err, base.MemberAccount, wallet.Currency)
	}
	//cache沒有資料就查sql db
	if base.MemberAccount == "" || wallet.Currency == "" {
		playerInfo := database.GetPlayerInfo(es.AddTraceMap(traceMap, sqlid.GetPlayerInfo.String()), account, currency, gameId)
		if playerInfo.MemberAccount == "" {
			es.Error("traceMap:%s ,get db error:%v ,base.MemberAccount:%s", traceMap, err, base.MemberAccount)
			selfDefine.ErrorCode = string(errorcode.UnknowError)
			return nil
		}
		database.SetPlayerBaseAndWallet(es.AddTraceMap(traceMap, redisid.SetPlayerInfoCache.String()), playerInfo)
		base = playerInfo.PlayerBase
		wallet = playerInfo.PlayerWallet
	}
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
