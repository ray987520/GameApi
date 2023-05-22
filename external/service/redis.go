package es

import (
	"time"

	"github.com/gomodule/redigo/redis"
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
		MaxIdle:     1, //空闲数
		IdleTimeout: 60 * time.Second,
		MaxActive:   10, //最大数
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
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
			return err
		},
	}
}

// 取單一redis string
func (pool *RedisPool) GetKey(key string) (value []byte, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	value, err = redis.Bytes(conn.Do("GET", key))
	return
}

// 設置redis string
func (pool *RedisPool) SetKey(key string, value []byte, ttlSecond int) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = conn.Do("SET", key, value)
	if ttlSecond > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSecond)
	}
	return
}

// 刪除redis keys
func (pool *RedisPool) DeleteKey(keys ...interface{}) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = redis.Int(conn.Do("DEL", keys...))
	return
}

// redis queue LPUSH
func (pool *RedisPool) LPushList(key string, value []byte) (err error) {
	conn := redisPool.Get()
	defer conn.Close()
	_, err = redis.Int(conn.Do("LPUSH", key, value))
	return
}

// 取redis client,需要使用MULTI之類時使用
func (pool *RedisPool) GetClient() redis.Conn {
	return redisPool.Get()
}

// 取多個redis string
func (pool *RedisPool) GetKeys(keys ...interface{}) (values [][]byte, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	datas, err := redis.Values(conn.Do("MGET", keys...))
	if datas != nil {
		for _, d := range datas {
			if d != nil {
				data := d.([]byte)
				values = append(values, data)
			}
		}
	}
	return
}

// redis INCR
func (pool *RedisPool) IncrKey(key string) (data int64, err error) {
	conn := redisPool.Get()
	defer conn.Close()
	data, err = redis.Int64(conn.Do("INCR", key))
	return
}
