package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
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
	Request  entity.AddGameLogRequest
	TraceMap string
}

// databinding&validate
func ParseAddGameLogRequest(traceMap string, r *http.Request) (request entity.AddGameLogRequest, err error) {
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

func (service *AddGameLogService) Exec() (data interface{}) {
	//catch panic
	defer es.PanicTrace(service.TraceMap)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//從token取出currency
	var currency string
	if _, currency, _ = parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true); currency == "" {
		return nil
	}

	//取匯率,後續計算統一幣值部分
	exchangeRate := currency2ExchangeRate(es.AddTraceMap(service.TraceMap, string(functionid.Currency2ExchangeRate)), &service.Request.BaseSelfDefine, currency)
	if exchangeRate == decimal.Zero {
		return nil
	}

	//insert gamelog
	addGameLog2Db(es.AddTraceMap(service.TraceMap, string(functionid.AddGameLog2Db)), &service.Request, exchangeRate)
	return nil
}

// 判斷connectToken是否正常
func isConnectTokenError(traceMap string, selfDefine *entity.BaseSelfDefine, token string) bool {
	if !isConnectTokenAlive(es.AddTraceMap(traceMap, string(functionid.IsConnectTokenAlive)), token) {
		selfDefine.ErrorCode = string(errorcode.Unauthorized)
		return true
	}
	return false
}

// currency轉成匯率
func currency2ExchangeRate(traceMap string, selfDefine *entity.BaseSelfDefine, currency string) decimal.Decimal {
	exchangeRate, err := database.GetCurrencyExchangeRate(es.AddTraceMap(traceMap, sqlid.GetCurrencyExchangeRate.String()), currency)
	if err != nil {
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return decimal.Zero
	}
	return exchangeRate
}

// 添加gamelog到Db“
func addGameLog2Db(traceMap string, data *entity.AddGameLogRequest, exchangeRate decimal.Decimal) {
	isOK := database.AddGameLog(es.AddTraceMap(traceMap, sqlid.AddGameLog.String()), data.GameLog, exchangeRate)
	if !isOK {
		data.ErrorCode = string(errorcode.UnknowError)
	}
}

func (service *AddGameLogService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}
