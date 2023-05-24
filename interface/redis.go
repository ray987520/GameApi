package iface

import "github.com/gomodule/redigo/redis"

//redis服務介面
type IRedis interface {
	GetKey(string, string) ([]byte, error)
	SetKey(string, string, []byte, int) error
	DeleteKey(string, ...interface{}) error
	LPushList(string, string, []byte) error
	GetClient(string) redis.Conn
	GetKeys(string, ...interface{}) ([][]byte, error)
	IncrKey(string, string) (int64, error)
}
