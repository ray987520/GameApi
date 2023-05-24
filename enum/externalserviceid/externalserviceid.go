package esid

type EsId string

// 列管所有external service,用於traceMap,對應不同套件所以特定編號
const (
	Aes128Encrypt         EsId = "AES1.1"
	Aes128Decrypt         EsId = "AES1.2"
	SqlSelect             EsId = "GORM1.1"
	SqlUpdate             EsId = "GORM1.2"
	SqlDelete             EsId = "GORM1.3"
	SqlCreate             EsId = "GORM1.4"
	SqlBatchCreate        EsId = "GORM1.5"
	SqlTransaction        EsId = "GORM1.6"
	JwtCreateConnectToken EsId = "JWT1.1"
	JwtValidConnectToken  EsId = "JWT1.2"
	RedisGetKey           EsId = "RDGO1.1"
	RedisSetKey           EsId = "RDGO1.2"
	RedisDeleteKey        EsId = "RDGO1.3"
	RedisLPushList        EsId = "RDGO1.4"
	RedisGetClient        EsId = "RDGO1.5"
	RedisGetKeys          EsId = "RDGO1.6"
	RedisIncrKey          EsId = "RDGO1.7"
	UuidGen               EsId = "SFID1.1"
)
