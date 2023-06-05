package database

import (
	"TestAPI/entity"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/enum/redisid"
	es "TestAPI/external/service"
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
)

// 注入redis client
func InitRedisPool(redis iface.IRedis) bool {
	redisPool = redis
	return true
}

// 取ConnectToken緩存
func GetConnectTokenCache(traceMap string, token string) string {
	key := fmt.Sprintf(gameTokenKey, token)
	value, err := redisPool.GetKey(es.AddTraceMap(traceMap, string(esid.RedisGetKey)), key)
	if err != nil {
		return ""
	}
	return string(value)
}

// 設置ConnectToken緩存
func SetConnectTokenCache(traceMap string, token string, ttl int) bool {
	key := fmt.Sprintf(gameTokenKey, token)
	err := redisPool.SetKey(es.AddTraceMap(traceMap, string(esid.RedisSetKey)), key, []byte("1"), ttl)
	return err == nil
}

// 清除玩家基本資料(不常異動)
func ClearPlayerInfoCache(traceMap string, data entity.AuthConnectTokenResponse) bool {
	baseKey := fmt.Sprintf(playerInfoKey, data.GameID, data.PlayerBase.Currency, data.MemberAccount)
	walletKey := fmt.Sprintf(playerWalletKey, data.PlayerBase.Currency, data.MemberAccount)
	err := redisPool.DeleteKey(es.AddTraceMap(traceMap, string(esid.RedisDeleteKey)), baseKey, walletKey)
	return err == nil
}

// 取玩家基本資料
func GetPlayerInfoCache(traceMap string, account, currency string, gameId int) (base entity.PlayerBase, wallet entity.PlayerWallet, err error) {
	baseKey := fmt.Sprintf(playerInfoKey, gameId, currency, account)
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	values, err := redisPool.GetKeys(es.AddTraceMap(traceMap, string(esid.RedisGetKeys)), baseKey, walletKey)
	if err != nil {
		return entity.PlayerBase{}, entity.PlayerWallet{}, err
	}
	//playerinfo/wallet的cache分開存放,所以應該取回2筆
	if len(values) != 2 {
		err = fmt.Errorf("GetPlayerInfoCache cache error ,count:%d", len(values))
		zaplog.Errorw(innererror.DBRedisError, innererror.FunctionNode, redisid.GetPlayerInfoCache, innererror.ErrorTypeNode, innererror.GetKeysPartialError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "baseKey", baseKey, "walletKey", walletKey)
		return entity.PlayerBase{}, entity.PlayerWallet{}, err
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), values[0], &base)
	if err != nil {
		return entity.PlayerBase{}, entity.PlayerWallet{}, err
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), values[1], &wallet)
	if err != nil {
		return entity.PlayerBase{}, entity.PlayerWallet{}, err
	}
	return base, wallet, nil
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
	byteData, err := es.JsonMarshal(es.AddTraceMap(traceMap, string(esid.JsonMarshal)), data)
	if err != nil {
		return false
	}
	err = redisPool.SetKey(es.AddTraceMap(traceMap, string(esid.RedisSetKey)), key, byteData, ttl)
	return err == nil
}

// 取單一將號
func GetGameSequenceNumber(traceMap string, prefix string) string {
	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	seqNo, err := redisPool.IncrKey(es.AddTraceMap(traceMap, string(esid.RedisIncrKey)), key)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s%d", prefix, seqNo)
}

// 取多將號
func GetGameSequenceNumbers(traceMap string, quantity int, prefix string) []string {
	key := fmt.Sprintf(gameSequenceNumberKey, prefix)
	//先預定數量,然後計算出連號
	seqNo, err := redisPool.IncrKeyBy(es.AddTraceMap(traceMap, string(esid.RedisIncrKeyBy)), key, quantity)
	if err != nil {
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
	value, err := redisPool.GetKey(es.AddTraceMap(traceMap, string(esid.RedisGetKey)), key)
	if err != nil {
		return ""
	}
	return string(value)
}

// 設置補單token緩存,ttl 1800秒
func SetFinishGameResultTokenCache(traceMap string, token string) bool {
	key := fmt.Sprintf(finishGameResultTokenKey, token)
	err := redisPool.SetKey(es.AddTraceMap(traceMap, string(esid.RedisSetKey)), key, []byte("1"), 1800)
	if err != nil {
		return false
	}
	return err == nil
}

// 取玩家錢包緩存
func GetPlayerWalletCache(traceMap string, account, currency string) (wallet entity.PlayerWallet, err error) {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	data, err := redisPool.GetKey(es.AddTraceMap(traceMap, string(esid.RedisGetKey)), walletKey)
	if err != nil {
		return entity.PlayerWallet{}, err
	}
	err = es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), data, &wallet)
	if err != nil {
		return entity.PlayerWallet{}, err
	}
	return wallet, nil
}

// 設置玩家錢包緩存
func SetPlayerWalletCache(traceMap string, account, currency string, data entity.PlayerWallet) bool {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	byteData, err := es.JsonMarshal(es.AddTraceMap(traceMap, string(esid.JsonMarshal)), data)
	if err != nil {
		return false
	}
	err = redisPool.SetKey(es.AddTraceMap(traceMap, string(esid.RedisSetKey)), walletKey, byteData, 10)
	return err == nil
}

// 清除玩家錢包緩存
func ClearPlayerWalletCache(traceMap string, currency, account string) bool {
	walletKey := fmt.Sprintf(playerWalletKey, currency, account)
	err := redisPool.DeleteKey(es.AddTraceMap(traceMap, string(esid.RedisDeleteKey)), walletKey)
	return err == nil
}
