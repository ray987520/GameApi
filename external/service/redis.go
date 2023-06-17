package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	redisConnectProtocol = "tcp"
	redisConnectServer   = mconfig.GetString("redis.connServer")
	maxRedisOpenConns    = mconfig.GetInt("redis.maxOpenConns")
	maxRedisIdleConns    = mconfig.GetInt("redis.maxIdleConns")
	maxRedisIdleSecond   = mconfig.GetDuration("redis.maxIdleSecond")
	redisDialError       = "redigo connection error:%v"
	redisPingError       = "redigo ping error:%v"
)

var redisPool *redis.Pool

type RedisPool struct {
}

// 取redis pool實例
func GetRedisPool() *RedisPool {
	RedisInit()
	return &RedisPool{}
}

// 初始化,連線redis
func RedisInit() {
	redisPool = &redis.Pool{
		MaxIdle:     maxRedisIdleConns, //空闲数
		IdleTimeout: maxRedisIdleSecond * time.Second,
		MaxActive:   maxRedisOpenConns, //最大数
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(redisConnectProtocol, redisConnectServer)
			if err != nil {
				err = fmt.Errorf(redisDialError, err)
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.ErrorInfoNode, err)
				return nil, err
			}
			/*TODO redis有password時應驗證
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			*/
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				err = fmt.Errorf(redisPingError, err)
				zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisInit, innererror.TraceNode, tracer.DefaultTraceId, innererror.ErrorInfoNode, err)
			}
			return err
		},
	}
}

// 取單一redis string
func (pool *RedisPool) GetKey(traceId string, key string) (value []byte) {
	conn := redisPool.Get()
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisGetKey, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key)
		return nil
	}
	return value
}

// 設置redis string
func (pool *RedisPool) SetKey(traceId string, key string, value []byte, ttlSecond int) (isOK bool) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisSetKey, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key, "value", string(value))
		return false
	}
	if ttlSecond > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSecond)
		if err != nil {
			zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisSetKey, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key, "ttlSecond", ttlSecond)
			return false
		}
	}
	return true
}

// 刪除redis keys
func (pool *RedisPool) DeleteKey(traceId string, keys ...interface{}) (isOK bool) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := redis.Int(conn.Do("DEL", keys...))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisDeleteKey, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "keys", keys)
		return false
	}
	return true
}

// redis queue LPUSH
func (pool *RedisPool) LPushList(traceId string, key string, value []byte) (isOK bool) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := redis.Int(conn.Do("LPUSH", key, value))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisLPushList, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key, "value", string(value))
		return false
	}
	return true
}

// 取redis client,需要使用MULTI之類時使用
func (pool *RedisPool) GetClient(traceId string) redis.Conn {
	return redisPool.Get()
}

// 取多個redis string
func (pool *RedisPool) GetKeys(traceId string, keys ...interface{}) (values [][]byte) {
	conn := redisPool.Get()
	defer conn.Close()
	datas, err := redis.Values(conn.Do("MGET", keys...))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisGetKeys, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "keys", keys)
		return nil
	}
	for _, d := range datas {
		if d != nil {
			data := d.([]byte)
			values = append(values, data)
		}
	}

	return values
}

// redis INCR
func (pool *RedisPool) IncrKey(traceId string, key string) (data int64) {
	conn := redisPool.Get()
	defer conn.Close()
	data, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisIncrKey, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key)
		return -1
	}
	return data
}

// redis INCRBY
func (pool *RedisPool) IncrKeyBy(traceId string, key string, count int) (data int64) {
	conn := redisPool.Get()
	defer conn.Close()
	data, err := redis.Int64(conn.Do("INCRBY", key, count))
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.RedisIncrKeyBy, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "key", key)
		return -1
	}
	return data
}
