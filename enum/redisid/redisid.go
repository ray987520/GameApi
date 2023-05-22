package redisid

import (
	"fmt"
	"strconv"
)

type RedisId int

// RedisId轉成string添加SQL標籤
func (redisId RedisId) String() string {
	id := strconv.Itoa(int(redisId))
	return fmt.Sprintf("RDS%s", id)
}

// 列管所有redis CRUD funcion,用於traceMap,調用的順序交錯所以編為流水號
const (
	Unknow RedisId = iota
	GetConnectTokenCache
	SetConnectTokenCache
	ClearPlayerInfoCache
	GetPlayerInfoCache
	SetPlayerInfoCache
	SetKey
	GetGameSequenceNumber
	GetGameSequenceNumbers
	GetFinishGameResultTokenCache
	SetFinishGameResultTokenCache
	GetPlayerWalletCache
	SetPlayerWalletCache
	IncrConnectTokenBetCount
	ClearPlayerWalletCache
)
