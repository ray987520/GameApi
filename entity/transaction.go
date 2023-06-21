package entity

import (
	es "TestAPI/external/service"

	"github.com/shopspring/decimal"
)

// 個人錢包轉至遊戲錢包httprequest
type RollOutRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	RollHistory
}

func (req *RollOutRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *RollOutRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 遊戲錢包出入帳history
type RollHistory struct {
	Token              string          `json:"connectToken" validate:"min=1"`
	TransID            string          `json:"transID" validate:"startswith=roll"`
	GameSequenceNumber string          `json:"gameSequenceNumber" validate:"min=1"`
	Amount             decimal.Decimal `json:"amount" validate:"number"`
	TakeAll            int             `json:"takeAll" validate:"oneof=0 1"`
	RollTime           string          `json:"rollTime" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
}

// 個人錢包轉至遊戲錢包responsedata
type RollOutResponse struct {
	Currency string          `json:"currencyKind"`
	Amount   decimal.Decimal `json:"amount"`
	Balance  decimal.Decimal `json:"balance"`
}

// 遊戲錢包轉至個人錢包httprequest
type RollInRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	RollInHistory
}

func (req *RollInRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *RollInRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 遊戲錢包轉至個人錢包responsedata
type RollInHistory struct {
	GameResult
	TransID string `json:"transID" validate:"startswith=rollIn-"`
}
