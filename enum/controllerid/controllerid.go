package controllerid

type ControllerId string

// 列管所有controller,用於traceMap,controller以小寫c_開頭
const (
	CreateGuestConnectToken ControllerId = "c_CreateGuestConnectToken"
	AuthConnectToken        ControllerId = "c_AuthConnectToken"
	UpdateTokenLocation     ControllerId = "c_UpdateTokenLocation"
	GetConnectTokenInfo     ControllerId = "c_GetConnectTokenInfo"
	GetConnectTokenAmount   ControllerId = "c_GetConnectTokenAmount"
	DelConnectToken         ControllerId = "c_DelConnectToken"
	GetSequenceNumber       ControllerId = "c_GetSequenceNumber"
	GetSequenceNumbers      ControllerId = "c_GetSequenceNumbers"
	RoundCheck              ControllerId = "c_RoundCheck"
	GameResult              ControllerId = "c_GameResult"
	FinishGameResult        ControllerId = "c_FinishGameResult"
	AddGameLog              ControllerId = "c_AddGameLog"
	OrderList               ControllerId = "c_OrderList"
	RollOut                 ControllerId = "c_RollOut"
	RollIn                  ControllerId = "c_RollIn"
	Settlement              ControllerId = "c_Settlement"
	Distribution            ControllerId = "c_Distribution"
	CurrencyList            ControllerId = "c_CurrencyList"
)
