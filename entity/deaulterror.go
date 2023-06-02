package entity

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
	traceHeader       = "traceid"
	requestTimeHeader = "requesttime"
	errorCodeHeader   = "errorcode"
)

// 預設錯誤httprequest
type DefaultError struct {
	BaseHttpRequest
	BaseSelfDefine
}
