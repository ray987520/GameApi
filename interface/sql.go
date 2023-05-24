package iface

//sql服務介面
type ISqlService interface {
	Select(string, interface{}, string, ...interface{}) error
	Update(string, string, ...interface{}) error
	Delete(string, string, ...interface{}) error
	Create(string, string, ...interface{}) error
	BatchCreate(string, string, interface{}, int) error
	Transaction(string, []string, ...[]interface{}) error
}
