package entity

import (
	es "TestAPI/external/service"
)

// 踢除令牌httprequest
type KickTokenRequest struct {
	BaseHttpRequest
	BaseSelfDefine
}

func (req *KickTokenRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *KickTokenRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 踢除令牌request body
type KickToken struct {
	AdminAccount string `json:"adminAccount" validate:"min=1"` //操作人員帳號
	Token        string `json:"token" validate:"min=1"`        //玩家登入令牌
	Reason       string `json:"reason" validate:"gt=0"`        //踢除令牌原因
}

// 踢除令牌responsedata
type KickTokenResponse struct {
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

// 確認令牌連線狀態httprequest
type IsTokenOnlineRequest struct {
	BaseHttpRequest
	BaseSelfDefine
}

func (req *IsTokenOnlineRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *IsTokenOnlineRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 確認令牌連線狀態request body
type IsTokenOnline struct {
	AdminAccount string `json:"adminAccount" validate:"min=1"` //操作人員帳號
	Token        string `json:"token" validate:"min=1"`        //玩家登入令牌
}

// 確認令牌連線狀態responsedata
type IsTokenOnlineResponse struct {
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}
