package entity

import (
	es "TestAPI/external/service"

	"github.com/shopspring/decimal"
)

// 活動結算httprequest
type SettlementRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Settlement
}

func (req *SettlementRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

func (req *SettlementRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

// 活動結算requestdata
type Settlement struct {
	ActivityIV         string          `json:"activityIV" validate:"min=1"`
	Rank               int             `json:"rank" validate:"min=1"`
	MemberID           int             `json:"memberID" validate:"gt=0"`
	GameSequenceNumber string          `json:"gameSequenceNumber"`
	Currency           string          `json:"currency" validate:"oneof=CNY JPY THB MMK VND MYR IDR USD USDT BDT PHP kVND kIDR SGD KRW INR HKD"`
	Prize              decimal.Decimal `json:"prize" validate:"number"`
}

// 活動派獎httprequest
type DistributionRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Distribution
}

func (req *DistributionRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

func (req *DistributionRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

// 活動派獎requestdata
type Distribution struct {
	ActivityIV  string          `json:"activityIV" validate:"min=1"`
	Rank        int             `json:"rank" validate:"min=1"`
	PrizePayout decimal.Decimal `json:"prizePayout" validate:"number"`
}
