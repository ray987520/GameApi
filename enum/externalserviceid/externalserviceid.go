package esid

type EsId string

// 列管所有external service,用於traceMap,External Service以es_開頭
const (
	Aes128Encrypt         EsId = "es_Aes128Encrypt"
	Aes128Decrypt         EsId = "es_Aes128Decrypt"
	SqlInit               EsId = "es_SqlInit"
	SqlSelect             EsId = "es_SqlSelect"
	SqlUpdate             EsId = "es_SqlUpdate"
	SqlDelete             EsId = "es_SqlDelete"
	SqlCreate             EsId = "es_SqlCreate"
	SqlBatchCreate        EsId = "es_SqlBatchCreate"
	SqlTransaction        EsId = "es_SqlTransaction"
	JwtCreateConnectToken EsId = "es_JwtCreateConnectToken"
	JwtValidConnectToken  EsId = "es_JwtValidConnectToken"
	RedisInit             EsId = "es_RedisInit"
	RedisGetKey           EsId = "es_RedisGetKey"
	RedisSetKey           EsId = "es_RedisSetKey"
	RedisDeleteKey        EsId = "es_RedisDeleteKey"
	RedisLPushList        EsId = "es_RedisLPushList"
	RedisGetClient        EsId = "es_RedisGetClient"
	RedisGetKeys          EsId = "es_RedisGetKeys"
	RedisIncrKey          EsId = "es_RedisIncrKey"
	RedisIncrKeyBy        EsId = "es_RedisIncrKeyBy"
	UuidGen               EsId = "es_UuidGen"
	ParseTime             EsId = "es_ParseTime"
	JsonMarshal           EsId = "es_JsonMarshal"
	JsonUnMarshal         EsId = "es_JsonUnMarshal"
	PanicTrace            EsId = "es_PanicTrace"
)
