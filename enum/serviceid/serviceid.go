package serviceid

type ServiceId string

// 列管所有domain service,用於traceMap,對應controller所以特定編號
const (
	ConcurrentEntry         ServiceId = "SV0.0"
	ConcurrentFetchJob      ServiceId = "SV0.1"
	DefaultError            ServiceId = "SV0.2"
	CreateGuestConnectToken ServiceId = "SV1.0"
	AuthConnectToken        ServiceId = "SV1.1"
	UpdateTokenLocation     ServiceId = "SV1.2"
	GetConnectTokenInfo     ServiceId = "SV1.3"
	GetConnectTokenAmount   ServiceId = "SV1.4"
	DelConnectToken         ServiceId = "SV1.5"
	GetSequenceNumber       ServiceId = "SV2.1"
	GetSequenceNumbers      ServiceId = "SV2.2"
	RoundCheck              ServiceId = "SV2.3"
	GameResult              ServiceId = "SV3.1"
	FinishGameResult        ServiceId = "SV3.2"
	AddGameLog              ServiceId = "SV3.3"
	OrderList               ServiceId = "SV4.1"
	RollOut                 ServiceId = "SV5.1"
	RollIn                  ServiceId = "SV5.2"
	Settlement              ServiceId = "SV6.1"
	Distribution            ServiceId = "SV6.2"
	CurrencyList            ServiceId = "SV7.1"
)
