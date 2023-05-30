package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/zaplog"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	redisConnectProtocol = mconfig.GetString("redis.connProtocol")
	redisConnectServer   = mconfig.GetString("redis.connServer")
	maxRedisOpenConns    = mconfig.GetInt("redis.maxOpenConns")
	maxRedisIdleConns    = mconfig.GetInt("redis.maxIdleConns")
	maxRedisIdleSecond   = mconfig.GetDuration("redis.maxIdleSecond")
)

var redisPool *redis.Pool

type RedisPool struct {
}

// 取redis pool實例
func GetRedisPool() *RedisPool {
	return &RedisPool{}
}

// 初始化,連線redis
func init() {
	redisPool = &redis.Pool{
		MaxIdle:     maxRedisIdleConns, //空闲数
		IdleTimeout: maxRedisIdleSecond * time.Second,
		MaxActive:   maxRedisOpenConns, //最大数
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(redisConnectProtocol, redisConnectServer)
			if err != nil {
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisInit, innererror.ErrorTypeNode, innererror.InitRedisError, innererror.ErrorInfoNode, err)
				return nil, err
			}
			/*
				if password != "" {
					if _, err := c.Do("AUTH", password); err != nil {
						c.Close()
						return nil, err
					}
				}
			*/
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisInit, innererror.ErrorTypeNode, innererror.InitRedisError, innererror.ErrorInfoNode, err)
			}
			return err
		},
	}
}

// 取單一redis string
func (pool *RedisPool) GetKey(traceMap string, key string) (value []byte, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	value, err = redis.Bytes(conn.Do("GET", key))
	zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisGetKey, innererror.ErrorTypeNode, innererror.GetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
	return
}

// 設置redis string
func (pool *RedisPool) SetKey(traceMap string, key string, value []byte, ttlSecond int) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = conn.Do("SET", key, value)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisSetKey, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key, "value", string(value))
		return
	}
	if ttlSecond > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSecond)
		if err != nil {
			zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisSetKey, innererror.ErrorTypeNode, innererror.SetKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key, "ttlSecond", ttlSecond)
			return
		}
	}
	return
}

// 刪除redis keys
func (pool *RedisPool) DeleteKey(traceMap string, keys ...interface{}) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = redis.Int(conn.Do("DEL", keys...))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisDeleteKey, innererror.ErrorTypeNode, innererror.DeleteKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "keys", keys)
		return
	}
	return
}

// redis queue LPUSH
func (pool *RedisPool) LPushList(traceMap string, key string, value []byte) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = redis.Int(conn.Do("LPUSH", key, value))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisLPushList, innererror.ErrorTypeNode, innererror.LPushListError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key, "value", string(value))
		return
	}
	return
}

// 取redis client,需要使用MULTI之類時使用
func (pool *RedisPool) GetClient(traceMap string) redis.Conn {
	return redisPool.Get()
}

// 取多個redis string
func (pool *RedisPool) GetKeys(traceMap string, keys ...interface{}) (values [][]byte, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	datas, err := redis.Values(conn.Do("MGET", keys...))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisGetKeys, innererror.ErrorTypeNode, innererror.GetKeysError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "keys", keys)
		return
	}
	for _, d := range datas {
		if d != nil {
			data := d.([]byte)
			values = append(values, data)
		}
	}

	return
}

// redis INCR
func (pool *RedisPool) IncrKey(traceMap string, key string) (data int64, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	data, err = redis.Int64(conn.Do("INCR", key))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisIncrKey, innererror.ErrorTypeNode, innererror.IncrKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return
	}
	return
}

// redis INCRBY
func (pool *RedisPool) IncrKeyBy(traceMap string, key string, count int) (data int64, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	data, err = redis.Int64(conn.Do("INCRBY", key, count))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisIncrKeyBy, innererror.ErrorTypeNode, innererror.IncrKeyError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "key", key)
		return
	}
	return
}
