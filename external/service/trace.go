package es

import (
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"strings"
)

// 組合traceMap
func AddTraceMap(originMap, newStep string) string {
	if originMap == "" {
		return newStep
	}
	newMap := []string{originMap, newStep}
	return strings.Join(newMap, "_")
}

// panic記錄,在必須停止服務且非預計的error狀態做最後記錄
func PanicTrace(traceMap string) {
	r := recover()
	if r == nil {
		return
	}
	zaplog.Errorw(innererror.PanicError, innererror.TraceNode, traceMap, "panic", r)
}
