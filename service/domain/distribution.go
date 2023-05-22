package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/redisid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DistributionService struct {
	Request  entity.DistributionRequest
	TraceMap string
}

// databinding&validate
func ParseDistributionRequest(traceMap string, r *http.Request) (request entity.DistributionRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &request)
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

func (service *DistributionService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}

	account, wallet, isOK := getDistributionWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetDistributionWallet)), &service.Request.BaseSelfDefine, service.Request.Distribution)
	if !isOK {
		return
	}
	//資料沒派彩過才派彩
	if !hasUnpayActivityDistribution(es.AddTraceMap(service.TraceMap, string(functionid.HasUnpayActivityDistribution)), &service.Request.BaseSelfDefine, service.Request.ActivityIV, service.Request.Rank) {
		return
	}
	if isOK := activityDistribution(es.AddTraceMap(service.TraceMap, string(functionid.ActivityDistribution)), &service.Request.BaseSelfDefine, service.Request.Distribution, wallet.WalletID); !isOK {
		return
	}
	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), wallet.Currency, account)
	return
}

// 取派彩的用戶錢包
func getDistributionWallet(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Distribution) (account string, wallet entity.PlayerWallet, isOK bool) {
	account, wallet, err := database.GetDistributionWallet(es.AddTraceMap(traceMap, sqlid.GetDistributionWallet.String()), data)
	if err != nil || wallet.Currency == "" {
		es.Error("traceMap:%s ,error:%v ,wallet.Currency:%s", traceMap, err, wallet.Currency)
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return
	}
	isOK = true
	return
}

// 是否有未派彩紀錄
func hasUnpayActivityDistribution(traceMap string, selfDefine *entity.BaseSelfDefine, activityIV string, rank int) bool {
	if !database.IsExistsUnpayActivityDistribution(es.AddTraceMap(traceMap, sqlid.IsExistsUnpayActivityDistribution.String()), activityIV, rank) {
		es.Error("traceMap:%s ,hasUnpayActivityDistribution error", traceMap)
		selfDefine.ErrorCode = string(errorcode.ActivityPayoutDone)
		return false
	}
	return true
}

// 活動派彩
func activityDistribution(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Distribution, walletID string) (isOK bool) {
	isOK = database.ActivityDistribution(es.AddTraceMap(traceMap, sqlid.ActivityDistribution.String()), data, walletID, es.LocalNow(8))
	if !isOK {
		es.Error("traceMap:%s ,activityDistribution error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *DistributionService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
