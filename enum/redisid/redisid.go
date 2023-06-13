package redisid

type RedisId string

// 列管所有redis CRUD funcion,用於traceMap,redis以rds_開頭
const (
	GetConnectTokenCache          RedisId = "rds_GetConnectTokenCache"
	SetConnectTokenCache          RedisId = "rds_SetConnectTokenCache"
	ClearPlayerInfoCache          RedisId = "rds_ClearPlayerInfoCache"
	GetPlayerInfoCache            RedisId = "rds_GetPlayerInfoCache"
	SetPlayerInfoCache            RedisId = "rds_SetPlayerInfoCache"
	SetKey                        RedisId = "rds_SetKey"
	GetGameSequenceNumber         RedisId = "rds_GetGameSequenceNumber"
	GetGameSequenceNumbers        RedisId = "rds_GetGameSequenceNumbers"
	GetFinishGameResultTokenCache RedisId = "rds_GetFinishGameResultTokenCache"
	SetFinishGameResultTokenCache RedisId = "rds_SetFinishGameResultTokenCache"
	GetPlayerWalletCache          RedisId = "rds_GetPlayerWalletCache"
	SetPlayerWalletCache          RedisId = "rds_SetPlayerWalletCache"
	IncrConnectTokenBetCount      RedisId = "rds_IncrConnectTokenBetCount"
	ClearPlayerWalletCache        RedisId = "rds_ClearPlayerWalletCache"
)
