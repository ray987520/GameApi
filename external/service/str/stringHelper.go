package str

import (
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"strconv"
)

// 封裝strconv.Atoi
func Atoi(traceId, input string) (data int, isOK bool) {
	data, err := strconv.Atoi(input)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, innererror.Atoi, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "input", input)
		return 0, false
	}
	return data, true
}

// 封裝strconv.Itoa
func Itoa(traceId string, input int) (data string) {
	return strconv.Itoa(input)
}
