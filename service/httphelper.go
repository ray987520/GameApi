package service

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/innererror"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	"net/http"
	"net/url"

	"moul.io/http2curl"
)

const (
	logResponse             = "Log Response"
	getHttpResponseFunction = "GetHttpResponse"
)

// 封裝Http Repsonse
func GetHttpResponse(code string, requestTime, traceId string, data interface{}) entity.BaseHttpResponse {
	//在不正常response code完全沒賦值時,預設為UnknowError
	if code == string(errorcode.Default) {
		code = string(errorcode.UnknowError)
	}
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
			TraceCode: traceId,
		},
	}
	zaplog.Infow(logResponse, innererror.FunctionNode, getHttpResponseFunction, innererror.TraceNode, traceId, "response", resp)
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
	return decodeData, nil
}

// 將http request內容轉成curl字串,方便log與復現
func HttpRequest2Curl(req *http.Request) (string, error) {
	curl, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return curl.String(), nil
}
