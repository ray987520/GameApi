package iface

//uuid服務介面
type IUuid interface {
	//gen a uuid
	Gen(string) string
}
