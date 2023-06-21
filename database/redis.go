package database

import (
	"TestAPI/entity"
	"TestAPI/enum/innererror"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"fmt"
	"strconv"
)

var redisPool iface.IRedis

// 訂定redis key格式
const (
	gameTokenKey             = "gameToken:%s"             //連線token,gameToken:[ConnectToken]
	playerInfoKey            = "player:%d_%s_%s"          //玩家基本資料(不常異動),player:[GameId]_[Currency]_[Account]
	playerWalletKey          = "wallet:%s_%s"             //玩家錢包(常異動),wallet:[Currency]_[Account]
	playerBetCountKey        = "betcount:%s"              //玩家連線的下注次數,betcount:[ConnectToken]
	gameSequenceNumberKey    = "gameSequenceNumber:%s"    //用來取將號的key,gameSequenceNumber:[Prefix]
	finishGameResultTokenKey = "finishGameResultToken:%s" //補單用token,finishGameResultToken:[ConnectToken]
	tokenDefault             = "1"
	cacheCountError          = "redis cache count error"
)

// 注入redis client
func InitRedisPool(redis iface.IRedis) bool {
	redisPool = redis
	return true
}

// 取ConnectToken緩存
func GetConnectTokenCache(traceId string, token string) string {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetConnectTokenCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("token", token))

	key := fmt.Sprintf(gameTokenKey, token)
	value := redisPool.GetKey(traceId, key)
	//底層錯誤
	if value == nil {
		return ""
	}

	return string(value)
}

// 設置ConnectToken緩存
func SetConnectTokenCache(traceId string, token string, ttl int) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.SetConnectTokenCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("token", token, "ttl", ttl))

	key := fmt.Sprintf(gameTokenKey, token)
	return redisPool.SetKey(traceId, key, []byte(tokenDefault), ttl)
}

// 清除玩家基本資料(不常異動)
func ClearPlayerInfoCache(traceId string, data entity.AuthConnectTokenResponse) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.ClearPlayerInfoCache, innererror.TraceNode, traceId, "data", data)

	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	return redisPool.DeleteKey(traceId, baseKey, walletKey)
}

// 取玩家基本資料
func GetPlayerInfoCache(traceId string, account, currency string, gameId int) (base entity.PlayerBase, wallet entity.PlayerWallet) {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetPlayerInfoCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("account", account, "currency", currency, "gameId", gameId))

	baseKey := fmt.Sprintf(playerInfoKey, gameId, currency, account)
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	values := redisPool.GetKeys(traceId, baseKey, walletKey)
	//底層錯誤
	if values == nil {
		return entity.PlayerBase{}, entity.PlayerWallet{}
	}

	//not expected cache count,playerinfo/wallet的cache分開存放,所以應該取回2筆
	if len(values) != 2 {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerInfoCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, cacheCountError, "baseKey", baseKey, "walletKey", walletKey, "len(values)", len(values)))
		return entity.PlayerBase{}, entity.PlayerWallet{}
	}

	isOK := es.JsonUnMarshal(traceId, values[0], &base)
	//json deserialize error
	if !isOK {
		return entity.PlayerBase{}, entity.PlayerWallet{}
	}

	isOK = es.JsonUnMarshal(traceId, values[1], &wallet)
	//json deserialize error
	if !isOK {
		return entity.PlayerBase{}, entity.PlayerWallet{}
	}

	return base, wallet
}

// 設置玩家基本資料
func SetPlayerInfoCache(traceId string, data entity.AuthConnectTokenResponse, token string) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.SetPlayerInfoCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("data", data))

	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	setKey(traceId, baseKey, data.PlayerBase, 0)
	betCountKey := fmt.Sprintf(playerBetCountKey, token)
	setKey(traceId, betCountKey, data.PlayerBetCount.BetCount, 0)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	setKey(traceId, walletKey, data.PlayerWallet, 10)
	return true
}

// 設置key,data為struct
func setKey(traceId string, key string, data interface{}, ttl int) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.SetKey, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("key", key, "data", data))

	byteData := es.JsonMarshal(traceId, data)
	//json serialize error
	if byteData == nil {
		return false
	}

	return redisPool.SetKey(traceId, key, byteData, ttl)
}

// 取單一將號
func GetGameSequenceNumber(traceId string, prefix string) string {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetGameSequenceNumber, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("prefix", prefix))

	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	seqNo := redisPool.IncrKey(traceId, key)
	//底層錯誤
	if seqNo == -1 {
		return ""
	}

	return fmt.Sprintf("%s%d", prefix, seqNo)
}

// 取多將號
func GetGameSequenceNumbers(traceId string, quantity int, prefix string) []string {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetGameSequenceNumbers, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("quantity", quantity, "prefix", prefix))

	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	//先預定數量,然後計算出連號
	seqNo := redisPool.IncrKeyBy(traceId, key, quantity)
	//底層錯誤
	if seqNo == -1 {
		return []string{}
	}

	seqNos := []string{}
	for i := 1; i <= quantity; i++ {
		seqNos = append(seqNos, strconv.FormatInt(seqNo-int64(quantity-i), 10))
	}
	return seqNos
}

// 取補單token緩存
func GetFinishGameResultTokenCache(traceId string, token string) string {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetFinishGameResultTokenCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("token", token))

	key := fmt.Sprintf(finishGameResultTokenKey, token)
	value := redisPool.GetKey(traceId, key)
	//底層錯誤
	if value == nil {
		return ""
	}

	return string(value)
}

// 設置補單token緩存,ttl 1800秒
func SetFinishGameResultTokenCache(traceId string, token string) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.SetFinishGameResultTokenCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("token", token))

	key := fmt.Sprintf(finishGameResultTokenKey, token)
	return redisPool.SetKey(traceId, key, []byte(tokenDefault), 1800)
}

// 取玩家錢包緩存
func GetPlayerWalletCache(traceId string, account, currency string) (wallet entity.PlayerWallet, isOK bool) {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.GetPlayerWalletCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("account", account, "currency", currency))

	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	data := redisPool.GetKey(traceId, walletKey)
	//redis no data
	if data == nil {
		return entity.PlayerWallet{}, false
	}

	//not expected data
	isOK = es.JsonUnMarshal(traceId, data, &wallet)
	if !isOK {
		return entity.PlayerWallet{}, false
	}

	return wallet, true
}

// 設置玩家錢包緩存
func SetPlayerWalletCache(traceId string, account, currency string, data entity.PlayerWallet) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.SetPlayerWalletCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("account", account, "currency", currency, "data", data))

	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	byteData := es.JsonMarshal(traceId, data)
	//json serialize error
	if byteData == nil {
		return false
	}

	return redisPool.SetKey(traceId, walletKey, byteData, 10)
}

// 清除玩家錢包緩存
func ClearPlayerWalletCache(traceId string, currency, account string) bool {
	zaplog.Infow(dbInfo, innererror.FunctionNode, redisid.ClearPlayerWalletCache, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("account", account, "currency", currency))

	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	return redisPool.DeleteKey(traceId, walletKey)
}
