package esid

type EsId string

// 列管所有external service,用於traceMap,對應不同套件所以特定編號
const (
	Aes128Encrypt EsId = "AES1.1"
	Aes128Decrypt
)
