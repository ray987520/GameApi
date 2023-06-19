package innererror

const (
	TraceNode     = "traceId"   //用於zaplog,traceId節點名稱
	FunctionNode  = "function"  //用於zaplog,function節點名稱
	ErrorInfoNode = "error"     //用於zaplog,error節點名稱
	ErrorTypeNode = "errorType" //用於zaplog,errorType節點名稱
	DataNode      = "data"      //用於zaplog,data節點名稱
	InfoNode      = "info"      //用於zaplog,info節點名稱
)

const (
	DBRedisError         = "redis error"            //用於zaplog,internal error type
	DBSqlError           = "sql error"              //用於zaplog,internal error type
	ServiceError         = "service error"          //用於zaplog,internal error type
	ExternalServiceError = "external service error" //用於zaplog,internal error type
	ValidRequestError    = "Bad HttpRequest"        //用於zaplog,internal error type
	ConfigError          = "config error"           //用於zaplog,internal error type
	PanicError           = "panic error"            //用於zaplog,internal error type
	MiddlewareError      = "middleware error"       //用於zaplog,internal error type
)

type JsonError string

// 列管internal json序列反序列化 error,用於zaplog分類
const (
	JsonMarshalError   JsonError = "Json Marshal error"
	JsonUnMarshalError JsonError = "Json UnMarshal error"
)

type TimeError string

// 列管internal time usage error,用於zaplog分類
const (
	TimeParseError TimeError = "Time Parse Error"
)

type DataTransferError string

// 列管internal data轉換 error,用於zaplog分類
const (
	StringToIntError DataTransferError = "Transfer String2Int Error"
)

type DomainError string

// 列管internal domain檢查 error,用於zaplog分類
const (
	BaseCheckError DomainError = "Request Validate Error"
)

type SonyflakeError string

// 列管sonyflake uuid error,用於zaplog分類
const (
	InitFlakeError SonyflakeError = "Init SonyFlake Error"
	GenUidError    SonyflakeError = "Uuid Gen Error"
)

type ViperError string

// 列管viper error,用於zaplog分類
const (
	ReadConfigError ViperError = "Viper Read Config Error"
)

type MConfigId string

// mconfig zaplog分類,cfg_開頭
const (
	MConfigInit           MConfigId = "cfg_init"
	MConfigGet            MConfigId = "cfg_Get"
	MConfigGetString      MConfigId = "cfg_GetString"
	MConfigGetInt         MConfigId = "cfg_GetInt"
	MConfigGetInt64       MConfigId = "cfg_GetInt64"
	MConfigGetDuration    MConfigId = "cfg_GetDuration"
	MConfigGetStringSlice MConfigId = "cfg_GetStringSlice"
)

type StringExtend string

// StringExtend zaplog分類,str_開頭
const (
	Atoi StringExtend = "str_Atoi"
	Iota StringExtend = "str_Iota"
)
