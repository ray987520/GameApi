package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/external/service/tracer"
	"net/http"

	"github.com/shopspring/decimal"
)

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
	traceHeader       = "traceid"
	requestTimeHeader = "requesttime"
	errorCodeHeader   = "errorcode"
)

type AddGameLogService struct {
	Request entity.AddGameLogRequest
}

// databinding&validate
func ParseAddGameLogRequest(traceId string, r *http.Request) (request entity.AddGameLogRequest) {
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

	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)

	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *AddGameLogService) Exec() (data interface{}) {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//從token取出currency
	var currency string
	if _, currency, _ = parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true); currency == "" {
		return nil
	}

	//取匯率,後續計算統一幣值部分
	exchangeRate := currency2ExchangeRate(&service.Request.BaseSelfDefine, currency)
	if exchangeRate == decimal.Zero {
		return nil
	}

	//insert gamelog
	isOK := addGameLog2Db(&service.Request, exchangeRate)
	if !isOK {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

// 判斷connectToken是否正常
func isConnectTokenError(selfDefine *entity.BaseSelfDefine, token string) bool {
	if !isConnectTokenAlive(selfDefine.TraceID, token) {
		selfDefine.ErrorCode = string(errorcode.Unauthorized)
		return true
	}
	return false
}

// currency轉成匯率
func currency2ExchangeRate(selfDefine *entity.BaseSelfDefine, currency string) decimal.Decimal {
	exchangeRate := database.GetCurrencyExchangeRate(selfDefine.TraceID, currency)
	if exchangeRate.Equal(decimal.Zero) {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return decimal.Zero
	}
	return exchangeRate
}

// 添加gamelog到Db“
func addGameLog2Db(data *entity.AddGameLogRequest, exchangeRate decimal.Decimal) bool {
	isOK := database.AddGameLog(data.TraceID, data.GameLog, exchangeRate)
	if !isOK {
		data.ErrorCode = string(errorcode.UnknowError)
	}
	return isOK
}

func (service *AddGameLogService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
