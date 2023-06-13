package serviceid

type ServiceId string

// 列管所有domain service,用於traceMap,service以s_開頭
const (
	ConcurrentEntry         ServiceId = "s_ConcurrentEntry"
	ConcurrentFetchJob      ServiceId = "s_ConcurrentFetchJob"
	DefaultError            ServiceId = "s_DefaultError"
	CreateGuestConnectToken ServiceId = "s_CreateGuestConnectToken"
	AuthConnectToken        ServiceId = "s_AuthConnectToken"
	UpdateTokenLocation     ServiceId = "s_UpdateTokenLocation"
	GetConnectTokenInfo     ServiceId = "s_GetConnectTokenInfo"
	GetConnectTokenAmount   ServiceId = "s_GetConnectTokenAmount"
	DelConnectToken         ServiceId = "s_DelConnectToken"
	GetSequenceNumber       ServiceId = "s_GetSequenceNumber"
	GetSequenceNumbers      ServiceId = "s_GetSequenceNumbers"
	RoundCheck              ServiceId = "s_RoundCheck"
	GameResult              ServiceId = "s_GameResult"
	FinishGameResult        ServiceId = "s_FinishGameResult"
	AddGameLog              ServiceId = "s_AddGameLog"
	OrderList               ServiceId = "s_OrderList"
	RollOut                 ServiceId = "s_RollOut"
	RollIn                  ServiceId = "s_RollIn"
	Settlement              ServiceId = "s_Settlement"
	Distribution            ServiceId = "s_Distribution"
	CurrencyList            ServiceId = "s_CurrencyList"
)
