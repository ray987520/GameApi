package entity

import (
	"net/http"

	"github.com/shopspring/decimal"
)

// 寫入賽果(拉霸)httprequest
type GameResultRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	GameResult
}

func (req *GameResultRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 寫入賽果(拉霸)requestdata
type GameResult struct {
	Token                    string          `json:"connectToken" validate:"min=1"`
	GameSequenceNumber       string          `json:"gameSequenceNumber" validate:"min=1"`
	CurrencyKindBet          decimal.Decimal `json:"currencyKindBet" validate:"number"`
	CurrencyKindWinLose      decimal.Decimal `json:"currencyKindWinLose" validate:"number"`
	CurrencyKindPayout       decimal.Decimal `json:"currencyKindPayout" validate:"number"`
	CurrencyKindContribution decimal.Decimal `json:"currencyKindContribution" validate:"number"`
	CurrencyKindJackPot      decimal.Decimal `json:"currencyKindJackPot" validate:"number"`
	SequenceID               string          `json:"sequenceID" validate:"min=1"`
	GameRoom                 int             `json:"gameRoom" validate:"min=1"`
	BetTime                  string          `json:"betTime" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
	ServerTime               string          `json:"serverTime" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
	FreeGame                 int             `json:"freeGame" validate:"oneof=0 1"`
	TurnTimes                int             `json:"turnTimes" validate:"min=1"`
	BetMode                  int             `json:"betMode" validate:"oneof=0 1 2 3"`
}

// 寫入賽果(拉霸)responsedata
type GameResultResponse struct {
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
	BetCount int             `json:"betCount"`
}

// 補寫賽果(捕魚)httprequest
type FinishGameResultRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	FinishGameResult
}

func (req *FinishGameResultRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

// 補寫賽果(捕魚)requestdata
type FinishGameResult struct {
	GameResult
	TransID string `json:"transID" validate:"startswith=rollOut-"`
}

// 寫遊戲紀錄httprequest
type AddGameLogRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	GameLog
}

func (req *AddGameLogRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

func (request *AddGameLogRequest) ReadHttpHeader(r *http.Request) {
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
}

// 寫遊戲紀錄requestdata
type GameLog struct {
	Token                    string          `json:"connectToken" validate:"min=1"`
	GameSequenceNumber       string          `json:"gameSequenceNumber" validate:"min=1"`
	SequenceID               string          `json:"sequenceID" validate:"min=1"`
	GameLog                  interface{}     `json:"gameLog"`
	Bet                      decimal.Decimal `json:"bet" validate:"number"`
	WinLose                  decimal.Decimal `json:"winLose" validate:"number"`
	Payout                   decimal.Decimal `json:"payout" validate:"number"`
	Contribution             decimal.Decimal `json:"contribution" validate:"number"`
	JackPot                  decimal.Decimal `json:"jackPot" validate:"number"`
	CurrencyKindBet          decimal.Decimal `json:"currencyKindBet" validate:"number"`
	CurrencyKindWinLose      decimal.Decimal `json:"currencyKindWinLose" validate:"number"`
	CurrencyKindPayout       decimal.Decimal `json:"currencyKindPayout" validate:"number"`
	CurrencyKindContribution decimal.Decimal `json:"currencyKindContribution" validate:"number"`
	CurrencyKindJackPot      decimal.Decimal `json:"currencyKindJackPot" validate:"number"`
	BetTime                  string          `json:"betTime" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
}
