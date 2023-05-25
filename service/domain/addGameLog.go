package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"io/ioutil"
	"net/http"

	"github.com/shopspring/decimal"
)

type AddGameLogService struct {
	Request  entity.AddGameLogRequest
	TraceMap string
}

// databinding&validate
func ParseAddGameLogRequest(traceMap string, r *http.Request) (request entity.AddGameLogRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.ErrorCode = string(errorcode.UnknowError)
		return
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
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

func (service *AddGameLogService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	var currency string
	if _, currency, _ = parseConnectToken(es.AddTraceMap(service.TraceMap, string(functionid.ParseConnectToken)), &service.Request.BaseSelfDefine, service.Request.Token, true); currency == "" {
		return
	}
	//取匯率,後續計算統一幣值部分
	exchangeRate := currency2ExchangeRate(es.AddTraceMap(service.TraceMap, string(functionid.Currency2ExchangeRate)), &service.Request.BaseSelfDefine, currency)
	if exchangeRate == decimal.Zero {
		return
	}
	addGameLog2Db(es.AddTraceMap(service.TraceMap, string(functionid.AddGameLog2Db)), &service.Request, exchangeRate)
	return
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
	if isOK := database.AddGameLog(es.AddTraceMap(traceMap, sqlid.AddGameLog.String()), data.GameLog, exchangeRate); !isOK {
		data.ErrorCode = string(errorcode.UnknowError)
	}
}

func (service *AddGameLogService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
