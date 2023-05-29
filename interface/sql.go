package iface

//sql服務介面
type ISqlService interface {
	Select(string, interface{}, string, ...interface{}) (int64, error)
	Update(string, string, ...interface{}) (int64, error)
	Delete(string, string, ...interface{}) (int64, error)
	Create(string, string, ...interface{}) (int64, error)
	BatchCreate(string, string, interface{}, int) (int64, error)
	Transaction(string, []string, ...[]interface{}) (int64, error)
}
