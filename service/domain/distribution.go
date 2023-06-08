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

type DistributionService struct {
	Request  entity.DistributionRequest
	TraceMap string
}

// databinding&validate
func ParseDistributionRequest(traceMap string, r *http.Request) (request entity.DistributionRequest, err error) {
	body, err := readHttpRequestBody(es.AddTraceMap(traceMap, string(functionid.ReadHttpRequestBody)), r, &request)
	if err != nil {
		return request, err
	}

	err = parseJsonBody(es.AddTraceMap(traceMap, string(functionid.ParseJsonBody)), body, &request)
	if err != nil {
		return request, err
	}

	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request, err
	}
	return request, nil
}

func (service *DistributionService) Exec() (data interface{}) {
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	account, wallet, isOK := getDistributionWallet(es.AddTraceMap(service.TraceMap, string(functionid.GetDistributionWallet)), &service.Request.BaseSelfDefine, service.Request.Distribution)
	if !isOK {
		return nil
	}

	//資料沒派彩過才派彩
	hasRecord := hasUnpayActivityDistribution(es.AddTraceMap(service.TraceMap, string(functionid.HasUnpayActivityDistribution)), &service.Request.BaseSelfDefine, service.Request.ActivityIV, service.Request.Rank)
	if !hasRecord {
		return nil
	}

	isOK = activityDistribution(es.AddTraceMap(service.TraceMap, string(functionid.ActivityDistribution)), &service.Request.BaseSelfDefine, service.Request.Distribution, wallet.WalletID)
	if !isOK {
		return nil
	}

	database.ClearPlayerWalletCache(es.AddTraceMap(service.TraceMap, redisid.ClearPlayerWalletCache.String()), wallet.Currency, account)
	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 取派彩的用戶錢包
func getDistributionWallet(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Distribution) (account string, wallet entity.PlayerWallet, isOK bool) {
	account, wallet, err := database.GetDistributionWallet(es.AddTraceMap(traceMap, sqlid.GetDistributionWallet.String()), data)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", entity.PlayerWallet{}, false
	}
	return account, wallet, true
}

// 是否有未派彩紀錄
func hasUnpayActivityDistribution(traceMap string, selfDefine *entity.BaseSelfDefine, activityIV string, rank int) (hasRecord bool) {
	hasRecord = database.IsExistsUnpayActivityDistribution(es.AddTraceMap(traceMap, sqlid.IsExistsUnpayActivityDistribution.String()), activityIV, rank)
	if !hasRecord {
		selfDefine.ErrorCode = string(errorcode.ActivityPayoutDone)
	}
	return hasRecord
}

// 活動派彩
func activityDistribution(traceMap string, selfDefine *entity.BaseSelfDefine, data entity.Distribution, walletID string) (isOK bool) {
	isOK = database.ActivityDistribution(es.AddTraceMap(traceMap, sqlid.ActivityDistribution.String()), data, walletID, es.LocalNow(8))
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *DistributionService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
