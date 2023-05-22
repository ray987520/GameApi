package tokenstatus

type tokenStatus int

// GameToken的token狀態,1:Actived 2:Deleted
const (
	Actived tokenStatus = iota + 1
	Deleted
)
