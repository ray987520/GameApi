package iface

import "github.com/gomodule/redigo/redis"

//redis服務介面
type IRedis interface {
	//redis get key,return ([]byte)value
	GetKey(string, string) []byte
	//redis set key and set key ttl,ttl:0 =  no expire
	SetKey(string, string, []byte, int) bool
	//redis del keys
	DeleteKey(string, ...interface{}) bool
	//redis lpush,use in mq
	LPushList(string, string, []byte) bool
	//get a redis client,use in pipe/multi
	GetClient(string) redis.Conn
	//redis get keys
	GetKeys(string, ...interface{}) [][]byte
	//redis incr
	IncrKey(string, string) int64
	//redis incrby
	IncrKeyBy(string, string, int) int64
}
