package iface

import "TestAPI/entity"

//job介面,service應實現
type IJob interface {
	Exec() interface{}
	GetBaseSelfDefine() entity.BaseSelfDefine
}
