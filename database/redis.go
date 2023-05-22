package database

import (
	"TestAPI/entity"
	"TestAPI/enum/innererror"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
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
)

// 注入redis client
func InitRedisPool(redis iface.IRedis) bool {
	redisPool = redis
	return true
}

// 取ConnectToken緩存
func GetConnectTokenCache(traceMap string, token string) string {
	key := fmt.Sprintf(gameTokenKey, token)
	value, err := redisPool.GetKey(key)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetConnectTokenCache, innererror.ErrorTypeNode, innererror.GetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return ""
	}
	return string(value)
}

// 設置ConnectToken緩存
func SetConnectTokenCache(traceMap string, token string, ttl int) bool {
	key := fmt.Sprintf(gameTokenKey, token)
	err := redisPool.SetKey(key, []byte("1"), ttl)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetConnectTokenCache, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return false
	}
	return err == nil
}

// 清除玩家基本資料(不常異動)
func ClearPlayerInfoCache(traceMap string, data entity.AuthConnectTokenResponse) bool {
	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	err := redisPool.DeleteKey(baseKey, walletKey)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.ClearPlayerInfoCache, innererror.ErrorTypeNode, innererror.DeleteKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "baseKey", baseKey, "walletKey", walletKey)
		return false
	}
	return true
}

// 取玩家基本資料
func GetPlayerInfoCache(traceMap string, account, currency string, gameId int) (base entity.PlayerBase, wallet entity.PlayerWallet, err error) {
	baseKey := fmt.Sprintf(playerInfoKey, gameId, currency, account)
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	values, err := redisPool.GetKeys(baseKey, walletKey)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerInfoCache, innererror.ErrorTypeNode, innererror.GetKeysError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "baseKey", baseKey, "walletKey", walletKey)
		return
	}
	if len(values) != 2 {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerInfoCache, innererror.ErrorTypeNode, innererror.GetKeysPartialError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "baseKey", baseKey, "walletKey", walletKey)
		return
	}
	json.Unmarshal(values[0], &base)
	json.Unmarshal(values[1], &wallet)

	return
}

// 設置玩家基本資料
func SetPlayerInfoCache(traceMap string, data entity.AuthConnectTokenResponse, token string) bool {
	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	setKey(es.AddTraceMap(traceMap, redisid.SetKey.String()), baseKey, data.PlayerBase, 0)
	betCountKey := fmt.Sprintf(playerBetCountKey, token)
	setKey(es.AddTraceMap(traceMap, redisid.SetKey.String()), betCountKey, data.PlayerBetCount.BetCount, 0)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	setKey(es.AddTraceMap(traceMap, redisid.SetKey.String()), walletKey, data.PlayerWallet, 10)
	return true
}

// 設置玩家基本資料跟錢包
func SetPlayerBaseAndWallet(traceMap string, data entity.AuthConnectTokenResponse) bool {
	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	setKey(es.AddTraceMap(traceMap, redisid.SetKey.String()), baseKey, data.PlayerBase, 0)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	setKey(es.AddTraceMap(traceMap, redisid.SetKey.String()), walletKey, data.PlayerWallet, 10)
	return true
}

// 設置key,data為struct
func setKey(traceMap string, key string, data interface{}, ttl int) bool {
	byteData, err := json.Marshal(data)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetKey, innererror.ErrorTypeNode, innererror.JsonMarshalError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return false
	}
	err = redisPool.SetKey(key, byteData, ttl)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetKey, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return false
	}
	return err == nil
}

// 取單一將號
func GetGameSequenceNumber(traceMap string, prefix string) string {
	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	seqNo, err := redisPool.IncrKey(key)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetGameSequenceNumber, innererror.ErrorTypeNode, innererror.IncrKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return ""
	}
	return fmt.Sprintf("%s%d", prefix, seqNo)
}

// 取多將號
func GetGameSequenceNumbers(traceMap string, quantity int, prefix string) []string {
	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	conn := redisPool.GetClient()
	defer conn.Close()
	//先預定數量,然後計算出連號
	seqNo, err := redis.Int64(conn.Do("INCRBY", key, quantity))
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetGameSequenceNumbers, innererror.ErrorTypeNode, innererror.IncrKeyByError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key, "quantity", quantity)
		return []string{}
	}
	seqNos := []string{}
	for i := 1; i <= quantity; i++ {
		seqNos = append(seqNos, strconv.FormatInt(seqNo-int64(quantity-i), 10))
	}
	return seqNos
}

// 取補單token緩存
func GetFinishGameResultTokenCache(traceMap string, token string) string {
	key := fmt.Sprintf(finishGameResultTokenKey, token)
	value, err := redisPool.GetKey(key)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetFinishGameResultTokenCache, innererror.ErrorTypeNode, innererror.GetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return ""
	}
	return string(value)
}

// 設置補單token緩存,ttl 1800秒
func SetFinishGameResultTokenCache(traceMap string, token string) bool {
	key := fmt.Sprintf(finishGameResultTokenKey, token)
	err := redisPool.SetKey(key, []byte("1"), 1800)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetFinishGameResultTokenCache, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return false
	}
	return err == nil
}

// 取玩家錢包緩存
func GetPlayerWalletCache(traceMap string, account, currency string) (wallet entity.PlayerWallet, err error) {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	data, err := redisPool.GetKey(walletKey)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerWalletCache, innererror.ErrorTypeNode, innererror.GetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletKey", walletKey)
		return
	}
	err = json.Unmarshal(data, &wallet)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerWalletCache, innererror.ErrorTypeNode, innererror.JsonUnMarshalError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return
	}
	return
}

// 設置玩家錢包緩存
func SetPlayerWalletCache(traceMap string, account, currency string, data entity.PlayerWallet) bool {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	byteData, err := json.Marshal(data)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetPlayerWalletCache, innererror.ErrorTypeNode, innererror.JsonMarshalError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return false
	}
	err = redisPool.SetKey(walletKey, byteData, 10)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.SetPlayerWalletCache, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletKey", walletKey)
		return false
	}
	return true
}

// 計算連線BetCount
func IncrConnectTokenBetCount(traceMap string, token string, betTimes int) (count int64) {
	betCountKey := fmt.Sprintf(playerBetCountKey, token)
	conn := redisPool.GetClient()
	defer conn.Close()
	count, err := redis.Int64(conn.Do("INCRBY", betCountKey, betTimes))
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.IncrConnectTokenBetCount, innererror.ErrorTypeNode, innererror.IncrKeyByError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "betCountKey", betCountKey, "betTimes", betTimes)
		count = 0
		return
	}
	return
}

// 清除玩家錢包緩存
func ClearPlayerWalletCache(traceMap string, currency, account string) bool {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	err := redisPool.DeleteKey(walletKey)
	if err != nil {
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.ClearPlayerWalletCache, innererror.ErrorTypeNode, innererror.DeleteKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletKey", walletKey)
		return false
	}
	return err == nil
}
