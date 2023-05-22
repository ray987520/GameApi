package iface

//sql服務介面
type ISqlService interface {
	Select(interface{}, string, ...interface{}) error
	Update(string, ...interface{}) error
	Delete(string, ...interface{}) error
	Create(string, ...interface{}) error
	BatchCreate(string, interface{}, int) error
	Transaction([]string, ...[]interface{}) error
}
