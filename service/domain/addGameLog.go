package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"net/http"

	"github.com/shopspring/decimal"
)

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
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

	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//validate request
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

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.Currency2ExchangeRate, innererror.TraceNode, service.Request.TraceID, innererror.DataNode, tracer.MergeMessage("exchangeRate", exchangeRate))

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
