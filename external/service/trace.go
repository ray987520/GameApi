package es

import (
	"fmt"
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

// 封裝error,輸出錯誤行數與錯誤內容
func Error(any interface{}, args ...interface{}) error {
	if any != nil {
		err := (error)(nil)
		switch any.(type) {
		case string:
			err = fmt.Errorf(any.(string), args...)
		case error:
			err = fmt.Errorf(any.(error).Error(), args...)
		default:
			err = fmt.Errorf("%v", err)
		}
		_, fn, line, _ := runtime.Caller(1)
		fmt.Printf("Error: [%s:%d] %v \n", fn, line, err)
		return err
	}
	return nil
}
