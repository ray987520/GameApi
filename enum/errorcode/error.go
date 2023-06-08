package errorcode

type ErrorCode string

// 對應文件與ErrorMessage.Code
const (
	Default                 ErrorCode = "" //預設沒有errorCode,如果輸出時仍然沒值也等於有錯誤
	Success                 ErrorCode = "0"
	ApiConfigError          ErrorCode = "101"
	SignInvalid             ErrorCode = "102"
	ApiFailed               ErrorCode = "103"
	UnderMaintenance        ErrorCode = "104"
	IpBlock                 ErrorCode = "105"
	ApiTimeout              ErrorCode = "106"
	BadParameter            ErrorCode = "201"
	AccountDuplicate        ErrorCode = "202"
	TransactionIDDuplicate  ErrorCode = "203"
	InsufficientBalance     ErrorCode = "204"
	AccountNoExist          ErrorCode = "205"
	GameNoExist             ErrorCode = "206"
	SequenceNumberDuplicate ErrorCode = "208"
	ActivityRankLost        ErrorCode = "209"
	ActivityBonusError      ErrorCode = "210"
	ActivityPayoutDone      ErrorCode = "211"
	ActivityRankError       ErrorCode = "212"
	Unauthorized            ErrorCode = "401"
	NotFound                ErrorCode = "404"
	PlaceBetError           ErrorCode = "701"
	SettleBetError          ErrorCode = "702"
	UnknowError             ErrorCode = "999"
)
