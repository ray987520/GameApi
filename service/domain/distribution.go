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
	"net/http"
)

type DistributionService struct {
	Request entity.DistributionRequest
}

// databinding&validate
func ParseDistributionRequest(traceId string, r *http.Request) (request entity.DistributionRequest) {
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

	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *DistributionService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	//取派彩的錢包
	account, wallet, isOK := getDistributionWallet(&service.Request.BaseSelfDefine, service.Request.Distribution)
	if !isOK {
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetDistributionWallet, innererror.TraceNode, service.Request.TraceID, "account", account, "wallet", wallet)

	//資料沒派彩過才派彩
	hasRecord := hasUnpayActivityDistribution(&service.Request.BaseSelfDefine, service.Request.ActivityIV, service.Request.Rank)
	if !hasRecord {
		return nil
	}

	//派彩到錢包
	isOK = activityDistribution(&service.Request.BaseSelfDefine, service.Request.Distribution, wallet.WalletID)
	if !isOK {
		return nil
	}

	//清除錢包cache以同步
	database.ClearPlayerWalletCache(service.Request.TraceID, wallet.Currency, account)

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 取派彩的用戶錢包
func getDistributionWallet(selfDefine *entity.BaseSelfDefine, data entity.Distribution) (account string, wallet entity.PlayerWallet, isOK bool) {
	account, wallet = database.GetDistributionWallet(selfDefine.TraceID, data)
	//get wallet error
	if wallet.WalletID == "" {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return "", entity.PlayerWallet{}, false
	}

	return account, wallet, true
}

// 是否有未派彩紀錄
func hasUnpayActivityDistribution(selfDefine *entity.BaseSelfDefine, activityIV string, rank int) (hasRecord bool) {
	hasRecord = database.IsExistsUnpayActivityDistribution(selfDefine.TraceID, activityIV, rank)
	//沒需派彩資料
	if !hasRecord {
		selfDefine.ErrorCode = string(errorcode.ActivityPayoutDone)
	}

	return hasRecord
}

// 活動派彩
func activityDistribution(selfDefine *entity.BaseSelfDefine, data entity.Distribution, walletID string) (isOK bool) {
	isOK = database.ActivityDistribution(selfDefine.TraceID, data, walletID, es.LocalNow(8))
	//派彩失敗
	if !isOK {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}

	return isOK
}

func (service *DistributionService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
