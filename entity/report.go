package entity

//取得歷史紀錄網址httprequest
type OrderListRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Token string `json:"connectToken" validate:"min=1"`
}

//取得歷史紀錄網址responsedata
type OrderListResponse struct {
	Url string `json:"url"`
}
