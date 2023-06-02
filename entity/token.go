package entity

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// 取得測試令牌httprequest
type CreateGuestConnectTokenRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Account  string `json:"account" validate:"min=3,max=32,acct"`
	Currency string `json:"currency" validate:"oneof=CNY JPY THB MMK VND MYR IDR USD USDT BDT PHP kVND kIDR SGD KRW INR HKD"`
	GameId   int    `json:"gameID" validate:"gt=0"`
}

// 連線令牌結構,Key:[gameId]_[currency]_[account] ExpitreTime:timestamp+600
type ConnectToken struct {
	Key         string `json:"key" validate:"gt=0"`
	ExpitreTime int64  `json:"expire" validate:"gt=0"`
}

// 解析令牌輸出account, currency , gameId
func (token *ConnectToken) Parse() (account, currency string, gameId int) {
	datas := strings.Split(token.Key, "_")
	gameId, err := strconv.Atoi(datas[0])
	if err != nil {
		return "", "", 0
	}
	return datas[2], datas[1], gameId
}

// 取得測試令牌responsedata
type CreateGuestConnectTokenResponse struct {
	Token string `json:"connectToken"`
}

// 令牌登入httprequest
type AuthConnectTokenRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	AuthConnectToken
}

func (req *AuthConnectTokenRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 令牌登入requestdata
type AuthConnectToken struct {
	Token string `json:"connectToken" validate:"min=1"`
	Ip    string `json:"clientIpAddress" validate:"min=1"`
}

// 令牌登入responsedata
type AuthConnectTokenResponse struct {
	PlayerBase
	PlayerBetCount
	PlayerWallet
}

// 玩家資訊-不常變動內容
type PlayerBase struct {
	PlatformID    int             `json:"platformID" gorm:"column:platformID"`
	ChannelID     int             `json:"channelID" gorm:"column:channelID"`
	PoolID        string          `json:"poolID" gorm:"column:poolID"`
	GameID        int             `json:"gameID" gorm:"column:gameID"`
	MemberID      int             `json:"memberID" gorm:"column:memberID"`
	GameAccount   string          `json:"gameAccount" gorm:"column:gameAccount"`
	MemberAccount string          `json:"memberAccount" gorm:"column:memberAccount"`
	Currency      string          `json:"currencyKind" gorm:"column:currencyKind"`
	Threshold     decimal.Decimal `json:"threshold" gorm:"column:threshold"`
	App           bool            `json:"app" gorm:"column:app"`
	Report        bool            `json:"report" gorm:"column:report"`
	GamePlat      string          `json:"gamePlat" gorm:"column:gamePlat"`
	RTP           int             `json:"RTP" gorm:"column:RTP"`
}

// 玩家資訊-注單總計
type PlayerBetCount struct {
	BetCount int `json:"betCount" gorm:"column:betCount"`
}

// 玩家資訊-錢包,currency會與PlayerBase重複,若同時取用需要在輸出時調整
type PlayerWallet struct {
	WalletID string          `json:"walletId,omitempty" gorm:"column:walletId"`
	Amount   decimal.Decimal `json:"currency" gorm:"column:currency"`
	Currency string          `json:"currencyKind,omitempty" gorm:"column:walletCurrency"`
}

// 變更令牌位置httprequest
type UpdateTokenLocationRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	UpdateTokenLocation
}

func (req *UpdateTokenLocationRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 變更令牌位置requestdata
type UpdateTokenLocation struct {
	Token    string `json:"connectToken" validate:"min=1"`
	Location int    `json:"location" validate:"gte=0"` //遊戲位置 (0:大廳or房號/機台號)
}

// 取得令牌資訊httprequest
type GetConnectTokenInfoRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Token string `json:"connectToken" validate:"min=1"`
}

// 取得令牌資訊responsedata
type GetConnectTokenInfoResponse struct {
	PlatformID  int    `json:"platformID"`
	ChannelID   int    `json:"channelID"`
	MemberID    int    `json:"memberID"`
	GameID      int    `json:"gameID"`
	GameAccount string `json:"gameAccount"`
	Currency    string `json:"currencyKind"`
	PlayerWallet
}

// 取得令牌餘額httprequest
type GetConnectTokenAmountRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Token string `json:"connectToken" validate:"min=1"`
}

// 令牌登出httprequest
type DelConnectTokenRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Token string `json:"connectToken" validate:"min=1"`
}

func (req *DelConnectTokenRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}
