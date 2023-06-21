package entity

import es "TestAPI/external/service"

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
)

// 預設錯誤httprequest
type DefaultError struct {
	BaseHttpRequest
	BaseSelfDefine
}

func (req *DefaultError) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}
