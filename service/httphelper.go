package service

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"net/url"
)

// 封裝Http Repsonse
func GetHttpResponse(code string, requestTime, traceCode string, data interface{}) entity.BaseHttpResponse {
	errorMessage := database.GetExternalErrorMessage(es.AddTraceMap("", sqlid.GetExternalErrorMessage.String()), code)
	if errorMessage == "" {
		errorMessage = database.GetExternalErrorMessage(es.AddTraceMap("", sqlid.GetExternalErrorMessage.String()), string(errorcode.UnknowError))
	}
	resp := entity.BaseHttpResponse{
		Data: data,
		Status: entity.HttpResponseStatus{
			Code:      code,
			Message:   errorMessage,
			DateTime:  requestTime,
			TraceCode: traceCode,
		},
	}
	return resp
}

// 對url編碼避免特殊字元
func UrlEncode(data string) string {
	return url.QueryEscape(data)
}

// 對url已編碼特殊字元解碼
func UrlDecode(data string) (string, error) {
	decodeData, err := url.QueryUnescape(data)
	if err != nil {
		return "", err
	}
	return decodeData, err
}
