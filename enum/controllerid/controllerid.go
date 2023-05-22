package controllerid

type ControllerId string

// 列管所有controller,用於traceMap,對應文件所以特定編號
const (
	CreateGuestConnectToken ControllerId = "CR1.0"
	AuthConnectToken        ControllerId = "CR1.1"
	UpdateTokenLocation     ControllerId = "CR1.2"
	GetConnectTokenInfo     ControllerId = "CR1.3"
	GetConnectTokenAmount   ControllerId = "CR1.4"
	DelConnectToken         ControllerId = "CR1.5"
	GetSequenceNumber       ControllerId = "CR2.1"
	GetSequenceNumbers      ControllerId = "CR2.2"
	RoundCheck              ControllerId = "CR2.3"
	GameResult              ControllerId = "CR3.1"
	FinishGameResult        ControllerId = "CR3.2"
	AddGameLog              ControllerId = "CR3.3"
	OrderList               ControllerId = "CR4.1"
	RollOut                 ControllerId = "CR5.1"
	RollIn                  ControllerId = "CR5.2"
	Settlement              ControllerId = "CR6.1"
	Distribution            ControllerId = "CR6.2"
	CurrencyList            ControllerId = "CR7.1"
)
