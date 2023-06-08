package es

import (
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"bytes"
	"runtime"
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

	zaplog.Errorw(innererror.PanicError, innererror.TraceNode, traceMap, "panic", PanicTraceDetail())
}

// panic的時候輸出完整stacktrace
func PanicTraceDetail() string {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, 4096) //限制在4KB內
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return string(stack)
}
