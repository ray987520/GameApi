package service

import (
	"TestAPI/database"
	"TestAPI/enum/controllerid"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"TestAPI/service/domain"
	"net/http"

	"golang.org/x/sync/syncmap"
)

const (
	logRequestModel = "log request model"
)

var (
	sqlDb       iface.ISqlService
	redisPool   iface.IRedis
	ResponseMap syncmap.Map //改用sync.Map,適合key不同的大量讀寫,需要GO1.19
)

// init DB底層跟接收response的map
func InitConcurrentService() {
	sqlDb = es.GetSqlDb()
	database.InitSqlWorker(sqlDb)
	redisPool = es.GetRedisPool()
	database.InitRedisPool(redisPool)
	//sync.Map不需要初始化
	//ResponseMap = make(map[string]chan entity.BaseHttpResponse)
	initDispatcher()
}

// 依照domainID分類初始化job資料
func fetchJob(traceId string, domainNo controllerid.ControllerId, r *http.Request) (job iface.IJob) {
	switch domainNo {
	case controllerid.CreateGuestConnectToken:
		request := domain.ParseCreateGuestConnectTokenRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseCreateGuestConnectTokenRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.CreateGuestConnectTokenService{
			Request: request,
		}
	case controllerid.AuthConnectToken:
		request := domain.ParseAuthConnectTokenRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseAuthConnectTokenRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.AuthConnectTokenService{
			Request: request,
		}
	case controllerid.UpdateTokenLocation:
		request := domain.ParseUpdateTokenLocationRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseUpdateTokenLocationRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.UpdateTokenLocationService{
			Request: request,
		}
	case controllerid.GetConnectTokenInfo:
		request := domain.ParseGetConnectTokenInfoRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseGetConnectTokenInfoRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.GetConnectTokenInfoService{
			Request: request,
		}
	case controllerid.GetConnectTokenAmount:
		request := domain.ParseGetConnectTokenAmountRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseGetConnectTokenAmountRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.GetConnectTokenAmountService{
			Request: request,
		}
	case controllerid.DelConnectToken:
		request := domain.ParseDelConnectTokenRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseDelConnectTokenRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.DelConnectTokenService{
			Request: request,
		}
	case controllerid.GetSequenceNumber:
		request := domain.ParseGetSequenceNumberRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseGetSequenceNumberRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.GetSequenceNumberService{
			Request: request,
		}
	case controllerid.GetSequenceNumbers:
		request := domain.ParseGetSequenceNumbersRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseGetSequenceNumbersRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.GetSequenceNumbersService{
			Request: request,
		}
	case controllerid.RoundCheck:
		request := domain.ParseRoundCheckRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseRoundCheckRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.RoundCheckService{
			Request: request,
		}
	case controllerid.GameResult:
		request := domain.ParseGameResultRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseGameResultRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.GameResultService{
			Request: request,
		}
	case controllerid.FinishGameResult:
		request := domain.ParseFinishGameResultRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseFinishGameResultRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.FinishGameResultService{
			Request: request,
		}
	case controllerid.AddGameLog:
		request := domain.ParseAddGameLogRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseAddGameLogRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.AddGameLogService{
			Request: request,
		}
	case controllerid.OrderList:
		request := domain.ParseOrderListRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseOrderListRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.OrderListService{
			Request: request,
		}
	case controllerid.RollOut:
		request := domain.ParseRollOutRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseRollOutRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.RollOutService{
			Request: request,
		}
	case controllerid.RollIn:
		request := domain.ParseRollInRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseRollInRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.RollInService{
			Request: request,
		}
	case controllerid.Settlement:
		request := domain.ParseSettlementRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseSettlementRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.SettlementService{
			Request: request,
		}
	case controllerid.Distribution:
		request := domain.ParseDistributionRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseDistributionRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.DistributionService{
			Request: request,
		}
	case controllerid.CurrencyList:
		request := domain.ParseCurrencyListRequest(traceId, r)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseCurrencyListRequest, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.CurrencyListService{
			Request: request,
		}

	default:
		request := domain.ParseDefaultError(traceId, r)
		request.ErrorCode = string(errorcode.UnknowError)
		zaplog.Infow(logRequestModel, innererror.FunctionNode, functionid.ParseDefaultError, innererror.TraceNode, traceId, innererror.DataNode, request.ToString())
		job = &domain.DefaultErrorService{
			Request: request,
		}
	}
	return job
}

// enqueue job
func Entry(traceId string, domainNo controllerid.ControllerId, r *http.Request) {
	job := fetchJob(traceId, domainNo, r)
	JobQueue <- job
}
