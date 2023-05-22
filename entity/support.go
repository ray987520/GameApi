package entity

import "github.com/shopspring/decimal"

//取得支援幣別httprequest
type CurrencyListRequest struct {
	BaseHttpRequest
	BaseSelfDefine
}

//取得支援幣別responsedata
type CurrencyListResponse struct {
	CurrencyID   int64           `json:"currencyKindID" gorm:"column:id"`
	Currency     string          `json:"currencyKind" gorm:"column:currency"`
	ExchangeRate decimal.Decimal `json:"exchangeRate" gorm:"column:exchangeRate"`
}
