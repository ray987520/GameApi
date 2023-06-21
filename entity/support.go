package entity

import (
	es "TestAPI/external/service"

	"github.com/shopspring/decimal"
)

// 取得支援幣別httprequest
type CurrencyListRequest struct {
	BaseHttpRequest
	BaseSelfDefine
}

func (req *CurrencyListRequest) ToString() string {
	data := es.JsonMarshal(req.TraceID, req)
	return string(data)
}

func (req *CurrencyListRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 取得支援幣別responsedata
type CurrencyListResponse struct {
	CurrencyID   int64           `json:"currencyKindID" gorm:"column:id"`
	Currency     string          `json:"currencyKind" gorm:"column:currency"`
	ExchangeRate decimal.Decimal `json:"exchangeRate" gorm:"column:exchangeRate"`
}
