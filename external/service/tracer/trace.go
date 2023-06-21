package tracer

import (
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

const (
	DefaultTraceId  = "init"                  //在init或某些沒traceId的情況給與default值,之後查詢時方便識別重要錯誤
	stackTraceStart = "/src/runtime/panic.go" //runtime stack trace要解析的起點
	stackTraceEnd   = "\ngoroutine "          //runtime stack trace要解析的終點
)

// panic記錄,在必須停止服務且非預計的error狀態做最後記錄
func PanicTrace(traceId string) {
	//嘗試回復panic的錯誤,如果沒錯誤就返回
	r := recover()
	if r == nil {
		return
	}
	//輸出解析runtime stacktrace
	zaplog.Errorw(innererror.PanicError, innererror.TraceNode, traceId, innererror.DataNode, panicTraceDetail())
}

// panic的時候輸出解析runtime stacktrace
func panicTraceDetail() string {
	s := []byte(stackTraceStart)
	e := []byte(stackTraceEnd)
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

// 合併資料成log內容字串
func MergeMessage(keyValues ...interface{}) string {
	//input資料長度不正確返回空值
	if len(keyValues) == 0 || len(keyValues)%2 == 1 {
		return ""
	}

	//成對處理資料
	var messages []string
	for i := 0; i < len(keyValues); i += 2 {
		messages = append(messages, fmt.Sprintf("%v:%v", keyValues[i], keyValues[i+1]))
	}

	return strings.Join(messages, ", ")
}
