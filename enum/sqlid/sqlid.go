package sqlid

type SqlId string

// 列管所有sql CRUD funcion,用於traceMap,以sql_開頭
const (
	GetExternalErrorMessage           SqlId = "sql_GetExternalErrorMessage"
	GetCurrencyExchangeRate           SqlId = "sql_GetCurrencyExchangeRate"
	GetPlayerInfo                     SqlId = "sql_GetPlayerInfo"
	AddConnectToken                   SqlId = "sql_AddConnectToken"
	UpdateTokenLocation               SqlId = "sql_UpdateTokenLocation"
	GetTokenAlive                     SqlId = "sql_GetTokenAlive"
	DeleteToken                       SqlId = "sql_DeleteToken"
	AddGameResultReCountWallet        SqlId = "sql_AddGameResultReCountWallet"
	AddGameResult                     SqlId = "sql_AddGameResult"
	GetFinishGameResultTokenAlive     SqlId = "sql_GetFinishGameResultTokenAlive"
	GetPlayerWallet                   SqlId = "sql_GetPlayerWallet"
	IsExistsTokenGameResult           SqlId = "sql_IsExistsTokenGameResult"
	IsExistsRollInHistory             SqlId = "sql_IsExistsRollInHistory"
	AddRollInHistory                  SqlId = "sql_AddRollInHistory"
	AddGameLog                        SqlId = "sql_AddGameLog"
	GetGameLanguage                   SqlId = "sql_GetGameLanguage"
	AddRollOutHistory                 SqlId = "sql_AddRollOutHistory"
	AddActivityRank                   SqlId = "sql_AddActivityRank"
	IsExistsUnpayActivityDistribution SqlId = "sql_IsExistsUnpayActivityDistribution"
	ActivityDistribution              SqlId = "sql_ActivityDistribution"
	GetDistributionWallet             SqlId = "sql_GetDistributionWallet"
	GetCurrencyList                   SqlId = "sql_GetCurrencyList"
	GetRoundCheckList                 SqlId = "sql_GetRoundCheckList"
	IsExistsRolloutHistory            SqlId = "sql_IsExistsRolloutHistory"
	GetAccountBetCount                SqlId = "sql_GetAccountBetCount"
	GetAccountRtp                     SqlId = "sql_GetAccountRtp"
)
