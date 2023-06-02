package iface

import "TestAPI/entity"

//job介面,service應實現
type IJob interface {
	//service執行內容
	Exec() interface{}
	//service的request自訂欄位區塊
	GetBaseSelfDefine() entity.BaseSelfDefine
}
