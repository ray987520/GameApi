package sqlid

import (
	"fmt"
	"strconv"
)

type SqlId int

// SqlId轉成string添加SQL標籤
func (sqlId SqlId) String() string {
	id := strconv.Itoa(int(sqlId))
	return fmt.Sprintf("SQL%s", id)
}

// 列管所有sql CRUD funcion,用於traceMap,調用的順序交錯所以編為流水號
const (
	Unknow SqlId = iota
	GetExternalErrorMessage
	GetCurrencyExchangeRate
	GetPlayerInfo
	AddConnectToken
	UpdateTokenLocation
	GetTokenAlive
	DeleteToken
	AddGameResultReCountWallet
	AddGameResult
	GetFinishGameResultTokenAlive
	GetPlayerWallet
	IsExistsTokenGameResult
	IsExistsRollInHistory
	AddRollInHistory
	AddGameLog
	GetGameLanguage
	AddRollOutHistory
	AddActivityRank
	IsExistsUnpayActivityDistribution
	ActivityDistribution
	GetDistributionWallet
	GetCurrencyList
	GetRoundCheckList
	IsExistsRolloutHistory
	GetAccountBetCount
	GetAccountRtp
)
