package iface

//sql服務介面
type ISqlService interface {
	//sql select,return rowcount
	Select(string, interface{}, string, ...interface{}) int64
	//sql update,return rowsaffected
	Update(string, string, ...interface{}) int64
	//sql delete,return rowsaffected
	Delete(string, string, ...interface{}) int64
	//sql insert,return rowsaffected
	Create(string, string, ...interface{}) int64
	//sql batch insert,return rowsaffected
	BatchCreate(string, string, interface{}, int) int64
	//sql transaction,return rowsaffected
	Transaction(string, []string, ...[]interface{}) int64
}
