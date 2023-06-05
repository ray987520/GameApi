package service

import (
	"TestAPI/database"
	"TestAPI/enum/controllerid"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/serviceid"
	es "TestAPI/external/service"
	iface "TestAPI/interface"
	"TestAPI/service/domain"
	"net/http"

	"golang.org/x/sync/syncmap"
)

var (
	sqlDb       iface.ISqlService
	redisPool   iface.IRedis
	ResponseMap syncmap.Map //改用sync.Map,適合key不同的大量讀寫,需要GO1.19
)

// init DB底層跟接收response的map
func init() {
	sqlDb = es.GetSqlDb()
	database.InitSqlWorker(sqlDb)
	redisPool = es.GetRedisPool()
	database.InitRedisPool(redisPool)
	//sync.Map不需要初始化
	//ResponseMap = make(map[string]chan entity.BaseHttpResponse)
}

// 依照domainID分類初始化job資料
func fetchJob(traceMap string, domainNo controllerid.ControllerId, r *http.Request) (job iface.IJob) {
	switch domainNo {
	case controllerid.CreateGuestConnectToken:
		request, _ := domain.ParseCreateGuestConnectTokenRequest(es.AddTraceMap(traceMap, string(functionid.ParseCreateGuestConnectTokenRequest)), r)
		job = &domain.CreateGuestConnectTokenService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.CreateGuestConnectToken)),
		}
	case controllerid.AuthConnectToken:
		request, _ := domain.ParseAuthConnectTokenRequest(es.AddTraceMap(traceMap, string(functionid.ParseAuthConnectTokenRequest)), r)
		job = &domain.AuthConnectTokenService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.AuthConnectToken)),
		}
	case controllerid.UpdateTokenLocation:
		request, _ := domain.ParseUpdateTokenLocationRequest(es.AddTraceMap(traceMap, string(functionid.ParseUpdateTokenLocationRequest)), r)
		job = &domain.UpdateTokenLocationService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.UpdateTokenLocation)),
		}
	case controllerid.GetConnectTokenInfo:
		request, _ := domain.ParseGetConnectTokenInfoRequest(es.AddTraceMap(traceMap, string(functionid.ParseGetConnectTokenInfoRequest)), r)
		job = &domain.GetConnectTokenInfoService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.GetConnectTokenInfo)),
		}
	case controllerid.GetConnectTokenAmount:
		request, _ := domain.ParseGetConnectTokenAmountRequest(es.AddTraceMap(traceMap, string(functionid.ParseGetConnectTokenAmountRequest)), r)
		job = &domain.GetConnectTokenAmountService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.GetConnectTokenAmount)),
		}
	case controllerid.DelConnectToken:
		request, _ := domain.ParseDelConnectTokenRequest(es.AddTraceMap(traceMap, string(functionid.ParseDelConnectTokenRequest)), r)
		job = &domain.DelConnectTokenService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.DelConnectToken)),
		}
	case controllerid.GetSequenceNumber:
		request, _ := domain.ParseGetSequenceNumberRequest(es.AddTraceMap(traceMap, string(functionid.ParseGetSequenceNumberRequest)), r)
		job = &domain.GetSequenceNumberService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.GetSequenceNumber)),
		}
	case controllerid.GetSequenceNumbers:
		request, _ := domain.ParseGetSequenceNumbersRequest(es.AddTraceMap(traceMap, string(functionid.ParseGetSequenceNumbersRequest)), r)
		job = &domain.GetSequenceNumbersService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.GetSequenceNumbers)),
		}
	case controllerid.RoundCheck:
		request, _ := domain.ParseRoundCheckRequest(es.AddTraceMap(traceMap, string(functionid.ParseRoundCheckRequest)), r)
		job = &domain.RoundCheckService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.RoundCheck)),
		}
	case controllerid.GameResult:
		request, _ := domain.ParseGameResultRequest(es.AddTraceMap(traceMap, string(functionid.ParseGameResultRequest)), r)
		job = &domain.GameResultService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.GameResult)),
		}
	case controllerid.FinishGameResult:
		request, _ := domain.ParseFinishGameResultRequest(es.AddTraceMap(traceMap, string(functionid.ParseFinishGameResultRequest)), r)
		job = &domain.FinishGameResultService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.FinishGameResult)),
		}
	case controllerid.AddGameLog:
		request, _ := domain.ParseAddGameLogRequest(es.AddTraceMap(traceMap, string(functionid.ParseAddGameLogRequest)), r)
		job = &domain.AddGameLogService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.AddGameLog)),
		}
	case controllerid.OrderList:
		request, _ := domain.ParseOrderListRequest(es.AddTraceMap(traceMap, string(functionid.ParseOrderListRequest)), r)
		job = &domain.OrderListService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.OrderList)),
		}
	case controllerid.RollOut:
		request, _ := domain.ParseRollOutRequest(es.AddTraceMap(traceMap, string(functionid.ParseRollOutRequest)), r)
		job = &domain.RollOutService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.RollOut)),
		}
	case controllerid.RollIn:
		request, _ := domain.ParseRollInRequest(es.AddTraceMap(traceMap, string(functionid.ParseRollInRequest)), r)
		job = &domain.RollInService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.RollIn)),
		}
	case controllerid.Settlement:
		request, _ := domain.ParseSettlementRequest(es.AddTraceMap(traceMap, string(functionid.ParseSettlementRequest)), r)
		job = &domain.SettlementService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.Settlement)),
		}
	case controllerid.Distribution:
		request, _ := domain.ParseDistributionRequest(es.AddTraceMap(traceMap, string(functionid.ParseDistributionRequest)), r)
		job = &domain.DistributionService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.Distribution)),
		}
	case controllerid.CurrencyList:
		request, _ := domain.ParseCurrencyListRequest(es.AddTraceMap(traceMap, string(functionid.ParseCurrencyListRequest)), r)
		job = &domain.CurrencyListService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.CurrencyList)),
		}

	default:
		request, _ := domain.ParseDefaultError(es.AddTraceMap(traceMap, string(functionid.ParseDefaultError)), r)
		request.ErrorCode = string(errorcode.UnknowError)
		job = &domain.DefaultErrorService{
			Request:  request,
			TraceMap: es.AddTraceMap(traceMap, string(serviceid.DefaultError)),
		}
	}
	return job
}

// enqueue job
func Entry(traceMap string, domainNo controllerid.ControllerId, r *http.Request) {
	job := fetchJob(es.AddTraceMap(traceMap, string(serviceid.ConcurrentFetchJob)), domainNo, r)
	JobQueue <- job
}
