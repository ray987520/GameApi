package rankstatus

type rankStatus int

// 對應ActivityRank.status,活動派發紀錄的狀態,0:Unpay 1:Payed
const (
	UnPay rankStatus = iota
	Payed
)
