package iface

import "github.com/gomodule/redigo/redis"

//redis服務介面
type IRedis interface {
	GetKey(string) ([]byte, error)
	SetKey(string, []byte, int) error
	DeleteKey(...interface{}) error
	LPushList(string, []byte) error
	GetClient() redis.Conn
	GetKeys(...interface{}) ([][]byte, error)
	IncrKey(string) (int64, error)
}
