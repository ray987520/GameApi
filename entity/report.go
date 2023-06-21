package entity

import es "TestAPI/external/service"

//取得歷史紀錄網址httprequest
type OrderListRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Token string `json:"connectToken" validate:"min=1"`
}

func (req *OrderListRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *OrderListRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

//取得歷史紀錄網址responsedata
type OrderListResponse struct {
	Url string `json:"url"`
}
