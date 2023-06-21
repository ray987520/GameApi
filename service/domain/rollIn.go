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
	"strings"

	"github.com/shopspring/decimal"
)

type RollInService struct {
	Request entity.RollInRequest
}

// databinding&validate
func ParseRollInRequest(traceId string, r *http.Request) (request entity.RollInRequest) {
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

	//validate request,transId因為model公用所以在這裡另外寫條件
	if !IsValid(traceId, request) || !strings.HasPrefix(request.TransID, "rollIn-") {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}

	return request
}

func (service *RollInService) Exec() (data interface{}) {
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	if isConnectTokenError(&service.Request.BaseSelfDefine, service.Request.Token) {
		return nil
	}

	//parse game token
	account, currency, _ := parseConnectToken(&service.Request.BaseSelfDefine, service.Request.Token, true)
	if account == "" {
		return nil
	}

	//check tollOut record existed
	hasRollOut, _ := hasRollOutHistory(&service.Request.BaseSelfDefine, service.Request.GameSequenceNumber)
	if !hasRollOut {
		return nil
	}

	/* *TOCHECK 理論上rollout amount應等於rollint bet
	if !rollOutAmount.Equal(service.Request.CurrencyKindBet) {
		service.Request.ErrorCode = string(errorcode.BadParameter)
		return nil
	}
	*/

	//get wallet
	wallet, isOK := getPlayerWallet(&service.Request.BaseSelfDefine, account, currency)
	if !isOK {
		return nil
	}

	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, functionid.GetPlayerWallet, innererror.TraceNode, service.Request.TraceID, innererror.DataNode, tracer.MergeMessage("wallet", wallet))

	//try add game result
	isAddGameResultOK := addGameResult(&service.Request.BaseSelfDefine, service.Request.GameResult)
	if !isAddGameResultOK {
		return nil
	}

	//try add rollIn record
	isAddRollHistoryOK := addRollInHistory(&service.Request.BaseSelfDefine, service.Request.GameResult, wallet)
	if !isAddRollHistoryOK {
		return nil
	}

	data = refreshWallet(&service.Request.BaseSelfDefine, account, currency, service.Request.Token, service.Request.TurnTimes)
	if data == nil {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return data
}

// 防呆,檢查要先有rollOut
func hasRollOutHistory(selfDefine *entity.BaseSelfDefine, seqNo string) (hasData bool, rollOutAmount decimal.Decimal) {
	hasData, rollOutAmount = database.IsExistsRolloutHistory(selfDefine.TraceID, seqNo)
	//get no rollOut history
	if !hasData {
		selfDefine.ErrorCode = string(errorcode.BadParameter)
		return false, decimal.Zero
	}

	return hasData, rollOutAmount
}

func (service *RollInService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
