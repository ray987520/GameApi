package database

import (
	"TestAPI/entity"
	"TestAPI/enum/innererror"
	"TestAPI/enum/rankstatus"
	"TestAPI/enum/rolltype"
	"TestAPI/enum/sqlid"
	"TestAPI/enum/tokenlocation"
	"TestAPI/enum/tokenstatus"
	es "TestAPI/external/service"
	"TestAPI/external/service/str"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	rollOutIdFormat   = "rollOut-%s"                         //rollOut transactionId Format
	rollInIdFormat    = "rollIn-%s"                          //rollIn transactionId Format
	unknowError       = "Unknow Error"                       //unknow error default message
	rowCountError     = "rowsAffected is not match expected" //非預期輸出筆數
	currencyError     = "unknow currency:%s"                 //未知幣別
	emptyErrorMessage = "get no error message"               //取出空錯誤訊息
	dataError         = "data is not match expected"         //not expected data
)

var sqlDb iface.ISqlService

// 注入sql client
func InitSqlWorker(db iface.ISqlService) bool {
	sqlDb = db
	return true
}

// 取對外輸出的錯誤訊息
func GetExternalErrorMessage(traceId string, code string) (errorMessage string) {
	sql := `SELECT message
	FROM [dbo].[ErrorMessage](nolock)
	WHERE code=?
	AND codeType=1
	AND langCode='en-US'`
	params := []interface{}{code}
	rowCount := sqlDb.Select(traceId, &errorMessage, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return unknowError
	}

	//非預期輸出筆數
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetExternalErrorMessage, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return unknowError
	}

	//取出空錯誤訊息
	if errorMessage == "" {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetExternalErrorMessage, innererror.TraceNode, traceId, innererror.ErrorInfoNode, emptyErrorMessage, "sql", sql, "params", params, "errorMessage", errorMessage)
		return unknowError
	}
	return errorMessage
}

// 取匯率
func GetCurrencyExchangeRate(traceId string, currency string) (exchangeRate decimal.Decimal) {
	sql := `SELECT [exchangeRate]
			FROM [Currency](nolock)
			WHERE [currency]=?`
	params := []interface{}{currency}
	rowCount := sqlDb.Select(traceId, &exchangeRate, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return decimal.Zero
	}

	//非預期輸出筆數
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params)
		return decimal.Zero
	}

	//異常匯率
	if exchangeRate.LessThanOrEqual(decimal.Zero) {
		err := fmt.Errorf(currencyError, currency)
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err, "exchangeRate", exchangeRate, "rowCount", rowCount)
		return decimal.Zero
	}
	return exchangeRate
}

// 取玩家資訊
func GetPlayerInfo(traceId string, account, currency string, gameId int) (data entity.AuthConnectTokenResponse) {
	sql := `SELECT ch.pfCode as platformID
					,ch.chanId as channelID
					,'' as poolID
					,? as gameID
					,acct.acctId as memberID
					,0 as RTP
					,'' as gameAccount
					,acct.account as memberAccount
					,mc.id as walletId
					,mc.currency as currencyKind
					,mc.currency as walletCurrency
					,mc.amount as currency
					,ch.threshold
					,ch.app
					,ch.report
					,ch.gamePlat
					,0 as betCount
				FROM Account as acct(nolock)
				JOIN Channel as ch(nolock)
					ON acct.chanId=ch.chanId
				JOIN ManCoin as mc(nolock)
					ON acct.account=mc.account
				WHERE acct.account=?
					AND mc.currency=?`
	params := []interface{}{gameId, account, currency}
	rowCount := sqlDb.Select(traceId, &data, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return entity.AuthConnectTokenResponse{}
	}

	//非預期輸出筆數
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerInfo, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return entity.AuthConnectTokenResponse{}
	}

	data.GameAccount = fmt.Sprintf("%d_%s", data.ChannelID, data.MemberAccount)
	data.BetCount = GetAccountBetCount(traceId, account)
	data.RTP = GetAccountRtp(traceId, account)
	//異常返水率
	if data.RTP == 0 {
		return entity.AuthConnectTokenResponse{}
	}

	return data
}

// 建立連線token
func AddConnectToken(traceId string, token, account, currency, ip string, gameId int, now time.Time) bool {
	sql := `INSERT INTO [dbo].[GameToken]
			([connectToken]
			,[gameId]
			,[currency]
			,[account]
			,[loginTime]
			,[ip]
			,[location]
			,[status])
			VALUES (?,?,?,?,?,?,?,?)`
	params := []interface{}{token, gameId, currency, account, now.Format(es.DbTimeFormat), ip, int(tokenlocation.Default), int(tokenstatus.Actived)}
	rowCount := sqlDb.Create(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddConnectToken, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 1
}

// 刷新連線token位置
func UpdateTokenLocation(traceId string, token string, location int) bool {
	sql := `UPDATE [dbo].[GameToken]
			SET [location] = ?
 			WHERE [connectToken]=?`
	params := []interface{}{location, token}
	rowCount := sqlDb.Update(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.UpdateTokenLocation, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 1
}

// 連線token是否存活
func GetTokenAlive(traceId string, token string) (alive bool) {
	//如果在cache裡有值,返回true
	if GetConnectTokenCache(traceId, token) == tokenDefault {
		return true
	}

	//cache裡沒資料,從sql db取出後塞到cahe
	sql := `SELECT 1 as alive
		FROM [GameToken](nolock)
		WHERE [connectToken]=?
		AND status=?`
	rowCount := sqlDb.Select(traceId, &alive, sql, token, int(tokenstatus.Actived))
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetTokenAlive, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "token", token, "status", int(tokenstatus.Actived), "rowCount", rowCount)
		return false
	}

	//if token alive, push cache
	if alive {
		SetConnectTokenCache(traceId, token, 0)
	}

	return alive
}

// 登出刪除連線token
func DeleteToken(traceId string, token string, deleteTime time.Time) bool {
	sql := `UPDATE [dbo].[GameToken]
			SET [status] = ?
				,deleteTime=?
 			WHERE [connectToken]=?`
	params := []interface{}{int(tokenstatus.Deleted), deleteTime.Format(es.DbTimeFormat), token}
	rowCount := sqlDb.Update(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.DeleteToken, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	//set cache ttl=1800second
	SetConnectTokenCache(traceId, token, 1800)

	return rowCount == 1
}

// 添加GameResult|RollHistory並更新錢包
func AddGameResultReCountWallet(tracId string, data entity.GameResult, wallet entity.PlayerWallet, now time.Time) bool {
	sql := []string{`INSERT INTO [dbo].[GameResult]
						([connectToken]
						,[gameSequenceNumber]
						,[currencyKindBet]
						,[currencyKindWinLose]
						,[currencyKindPayout]
						,[currencyKindContribution]
						,[currencyKindJackPot]
						,[sequenceID]
						,[gameRoom]
						,[betTime]
						,[serverTime]
						,[freeGame]
						,[turnTimes]
						,[betMode])
					VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?);`,
		`SELECT 1
					FROM ManCoin WITH(HOLDLOCK)
					WHERE id=?`,
		`UPDATE ManCoin
					SET amount=amount+?
					WHERE id=?`}
	params := [][]interface{}{}
	betTime, isOK := es.ParseTime(tracId, es.ApiTimeFormat, data.BetTime)
	//parse time error
	if !isOK {
		return false
	}

	serverTime, isOK := es.ParseTime(tracId, es.ApiTimeFormat, data.ServerTime)
	//parse time error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{data.Token, data.GameSequenceNumber, data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout,
		data.CurrencyKindContribution, data.CurrencyKindJackPot, data.SequenceID, data.GameRoom, betTime.Format(es.DbTimeFormat), serverTime.Format(es.DbTimeFormat),
		data.FreeGame, data.TurnTimes, data.BetMode})
	walletId, isOK := str.Atoi(tracId, wallet.WalletID)
	//convert error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.CurrencyKindPayout.Sub(data.CurrencyKindBet), walletId})
	rowCount := sqlDb.Transaction(tracId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 3 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.TraceNode, tracId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 3
}

// 添加賽果
func AddGameResult(traceId string, data entity.GameResult) bool {
	sql := `INSERT INTO [dbo].[GameResult]
			([connectToken]
			,[gameSequenceNumber]
			,[currencyKindBet]
			,[currencyKindWinLose]
			,[currencyKindPayout]
			,[currencyKindContribution]
			,[currencyKindJackPot]
			,[sequenceID]
			,[gameRoom]
			,[betTime]
			,[serverTime]
			,[freeGame]
			,[turnTimes]
			,[betMode])
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	params := []interface{}{}
	betTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, data.BetTime)
	//parse time error
	if !isOK {
		return false
	}

	serverTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, data.ServerTime)
	//parse time error
	if !isOK {
		return false
	}

	params = append(params, data.Token, data.GameSequenceNumber, data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout,
		data.CurrencyKindContribution, data.CurrencyKindJackPot, data.SequenceID, data.GameRoom, betTime.Format(es.DbTimeFormat), serverTime.Format(es.DbTimeFormat),
		data.FreeGame, data.TurnTimes, data.BetMode)
	rowCount := sqlDb.Create(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResult, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 1
}

// 補單token是否存活
func GetFinishGameResultTokenAlive(traceId string, token string) (alive bool) {
	//目前設計補單用token只會存在於cache,且有特定key format
	return GetFinishGameResultTokenCache(traceId, token) == tokenDefault
}

// 取玩家錢包
func GetPlayerWallet(traceId string, account, currency string) (data entity.PlayerWallet, isOK bool) {
	data, isOK = GetPlayerWalletCache(traceId, account, currency)
	//if get cache success,return data
	if isOK {
		return data, isOK
	}

	sql := `SELECT [id] as walletId
					,[amount] as currency
					,[currency] as walletCurrency
				FROM [ManCoin]
				WHERE account=?
				AND currency=?`
	rowCount := sqlDb.Select(traceId, &data, sql, account, currency)
	//底層錯誤
	if rowCount == -1 {
		return entity.PlayerWallet{}, false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerWallet, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "account", account, "currency", currency, "rowCount", rowCount)
		return entity.PlayerWallet{}, false
	}

	//currency error
	if data.Currency == "" {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerWallet, innererror.TraceNode, traceId, innererror.ErrorInfoNode, fmt.Sprintf(currencyError, data.Currency))
		return entity.PlayerWallet{}, false
	}

	//if get wallet success, push cache
	SetPlayerWalletCache(traceId, account, currency, data)

	return data, true
}

// 以connectToken/將號取GameResult是否存在
func IsExistsTokenGameResult(traceId string, token, gameSeqNo string) (data bool) {
	sql := `SELECT 1
			FROM [GameResult](nolock)
			WHERE connectToken=?
			AND gameSequenceNumber=?`
	rowCount := sqlDb.Select(traceId, &data, sql, token, gameSeqNo)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount == 0 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsTokenGameResult, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "token", token, "gameSeqNo", gameSeqNo, "rowCount", rowCount)
		return false
	}

	return data
}

// 以connectToken/將號取RollInHistory是否存在
func IsExistsRollInHistory(traceId string, token, gameSeqNo string) (data bool) {
	rollInId := fmt.Sprintf(rollInIdFormat, gameSeqNo)
	sql := `SELECT 1
			FROM RollHistory(nolock)
			WHERE connectToken=?
			AND transId=?`
	rowCount := sqlDb.Select(traceId, &data, sql, token, rollInId)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount == 0 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRollInHistory, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "token", token, "rollInId", rollInId, "rowCount", rowCount)
		return false
	}

	return data
}

// RollIn並更新錢包
func AddRollInHistory(traceId string, data entity.GameResult, wallet entity.PlayerWallet, now time.Time) bool {
	sql := []string{`INSERT INTO [dbo].[RollHistory]
           ([connectToken]
           ,[transId]
           ,[gameSequenceNumber]
           ,[amount]
           ,[currency]
		   ,[rollType]
           ,[rollTime])
     	VALUES (?,?,?,?,?,?,?)`,
		`SELECT 1
		FROM ManCoin WITH(HOLDLOCK)
		WHERE id=?`,
		`UPDATE ManCoin
		SET amount=amount+?
		WHERE id=?`}
	params := [][]interface{}{}
	rollInId := fmt.Sprintf(rollInIdFormat, data.GameSequenceNumber)
	params = append(params, []interface{}{data.Token, rollInId, data.GameSequenceNumber, data.CurrencyKindPayout, wallet.Currency, int(rolltype.RollIn), now.Format(es.DbTimeFormat)})

	walletId, isOK := str.Atoi(traceId, wallet.WalletID)
	//convert error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.CurrencyKindPayout, walletId})
	rowCount := sqlDb.Transaction(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 3 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollInHistory, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 3
}

// 添加遊戲紀錄
func AddGameLog(traceId string, data entity.GameLog, exchangeRate decimal.Decimal) bool {
	sql := `INSERT INTO [dbo].[GameLog]
				([connectToken]
				,[gameSequenceNumber]
				,[sequenceID]
				,[gameLog]
				,[bet]
				,[winLose]
				,[payOut]
				,[contribution]
				,[jackPot]
				,[currencyKindBet]
				,[currencyKindWinLose]
				,[currencyKindPayout]
				,[currencyKindContribution]
				,[currencyKindJackPot]
				,[betTime])
			VALUES
				(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	params := []interface{}{}
	betTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, data.BetTime)
	//parse time error
	if !isOK {
		return false
	}

	params = append(params, data.Token, data.GameSequenceNumber, data.SequenceID, data.GameLog, data.CurrencyKindBet.Mul(exchangeRate), data.CurrencyKindWinLose.Mul(exchangeRate), data.CurrencyKindPayout.Mul(exchangeRate), data.CurrencyKindContribution.Mul(exchangeRate), data.CurrencyKindJackPot.Mul(exchangeRate), data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout, data.CurrencyKindContribution, data.CurrencyKindJackPot, betTime.Format(es.DbTimeFormat))
	rowCount := sqlDb.Create(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameLog, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 1
}

// 取遊戲語系
func GetGameLanguage(traceId string, gameId int) (data string) {
	sql := `SELECT lang
			FROM Game(nolock)
			WHERE gameId=?`
	rowCount := sqlDb.Select(traceId, &data, sql, gameId)
	//底層錯誤
	if rowCount == -1 {
		return ""
	}

	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetGameLanguage, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "gameId", gameId, "rowCount", rowCount)
		return ""
	}
	return data
}

// RollOut並更新錢包
func AddRollOutHistory(traceId string, data entity.RollHistory, wallet entity.PlayerWallet) bool {
	sql := []string{`INSERT INTO [dbo].[RollHistory]
           ([connectToken]
           ,[transId]
           ,[gameSequenceNumber]
           ,[amount]
           ,[currency]
		   ,[rollType]
           ,[rollTime])
     	VALUES (?,?,?,?,?,?,?)`,
		`SELECT 1
		FROM ManCoin WITH(HOLDLOCK)
		WHERE id=?`,
		`UPDATE ManCoin
		SET amount=amount+?
		WHERE id=?`}
	params := [][]interface{}{}
	rollTime, isOK := es.ParseTime(traceId, es.ApiTimeFormat, data.RollTime)
	//parse time error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{data.Token, data.TransID, data.GameSequenceNumber, data.Amount, wallet.Currency, int(rolltype.RollOut), rollTime.Format(es.DbTimeFormat)})
	walletId, isOK := str.Atoi(traceId, wallet.WalletID)
	//convert error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.Amount.Neg(), walletId})
	rowCount := sqlDb.Transaction(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 3 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollOutHistory, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 3
}

// 添加活動未派彩紀錄
func AddActivityRank(traceId string, data entity.Settlement) bool {
	sql := `INSERT INTO [dbo].[ActivityRank]
				([activityIV]
				,[rank]
				,[memberId]
				,[gameSequenceNumber]
				,[currency]
				,[prize]
				,[status])
			VALUES
				(?,?,?,?,?,?,?)`
	params := []interface{}{}
	params = append(params, data.ActivityIV, data.Rank, data.MemberID, data.GameSequenceNumber, data.Currency, data.Prize, int(rankstatus.UnPay))
	rowCount := sqlDb.Create(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddActivityRank, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 1
}

// 取是否有對應未派彩紀錄
func IsExistsUnpayActivityDistribution(traceId string, activityIV string, rank int) (hasData bool) {
	sql := `SELECT 1
			FROM [dbo].[ActivityRank]
			WHERE [activityIV]=?
			AND [rank]=?
			AND status=?`
	params := []interface{}{}
	params = append(params, activityIV, rank, int(rankstatus.UnPay))
	rowCount := sqlDb.Select(traceId, &hasData, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsUnpayActivityDistribution, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return hasData
}

// 派彩並更新錢包
func ActivityDistribution(traceId string, data entity.Distribution, walletID string, now time.Time) bool {
	sql := []string{`UPDATE [dbo].[ActivityRank]
			SET [prizePayout]=?,[payTime]=?,[status]=?
			WHERE [activityIV]=?
			AND [rank]=?
			AND status=?`,
		`SELECT 1
			FROM ManCoin WITH(HOLDLOCK)
			WHERE id=?`,
		`UPDATE ManCoin
			SET amount=amount+?
			WHERE id=?`}
	params := [][]interface{}{}
	params = append(params, []interface{}{data.PrizePayout, now, int(rankstatus.Payed), data.ActivityIV, data.Rank, int(rankstatus.UnPay)})
	walletId, isOK := str.Atoi(traceId, walletID)
	//convert error
	if !isOK {
		return false
	}

	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.PrizePayout, walletId})
	rowCount := sqlDb.Transaction(traceId, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false
	}

	//not expected rowCount
	if rowCount != 3 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.ActivityDistribution, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}

	return rowCount == 3
}

// 取派彩對象玩家錢包
func GetDistributionWallet(traceId string, data entity.Distribution) (account string, wallet entity.PlayerWallet) {
	temp := struct {
		Account  string
		Currency string
	}{}
	sql := `SELECT A.account,B.currency
		FROM Account(nolock) as A
		JOIN ActivityRank(nolock) as B
			ON A.acctId=B.memberId
		WHERE B.[activityIV]=?
		AND B.[rank]=?`
	rowCount := sqlDb.Select(traceId, &temp, sql, data.ActivityIV, data.Rank)
	//底層錯誤
	if rowCount == -1 {
		return "", entity.PlayerWallet{}
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetDistributionWallet, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "ActivityIV", data.ActivityIV, "Rank", data.Rank, "rowCount", rowCount)
		return "", entity.PlayerWallet{}
	}

	wallet, isOK := GetPlayerWallet(traceId, temp.Account, temp.Currency)
	//get wallet error
	if !isOK {
		return temp.Account, entity.PlayerWallet{}
	}

	return temp.Account, wallet
}

// 取支援的Currency清單
func GetCurrencyList(traceId string) (list []entity.CurrencyListResponse) {
	sql := `SELECT [id]
				,[currency]
				,[exchangeRate]
			FROM [Currency]`
	rowCount := sqlDb.Select(traceId, &list, sql)
	//底層錯誤
	if rowCount == -1 {
		return nil
	}

	//not expected rowCount
	if rowCount == 0 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyList, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "rowCount", rowCount)
		return nil
	}
	return list
}

// 查RollOut對應GameResult/RollIn
func GetRoundCheckList(traceId string, fromDate, toDate string) (list []entity.RoundCheckToken) {
	rollTimeStart, isOK := es.ParseTime(traceId, es.ApiTimeFormat, fromDate)
	//parse time error
	if !isOK {
		return nil
	}

	rollTimeEnd, isOK := es.ParseTime(traceId, es.ApiTimeFormat, toDate)
	//parse time error
	if !isOK {
		return nil
	}

	sql := `SELECT RO.connectToken,RO.gameSequenceNumber
			FROM RollHistory(nolock) as RO
			LEFT JOIN GameResult(nolock) as GR
			ON RO.connectToken=GR.connectToken AND RO.gameSequenceNumber=GR.gameSequenceNumber
			LEFT JOIN RollHistory(nolock) as RI
			ON RO.connectToken=RI.connectToken AND RO.gameSequenceNumber=RI.gameSequenceNumber AND RI.rollType=2
			WHERE RO.rollTime BETWEEN ? AND ? 
			AND RO.rollType=?
			AND (GR.id IS NULL OR RI.id IS NULL)`
	params := []interface{}{}
	params = append(params, rollTimeStart.Format(es.DbTimeFormat), rollTimeEnd.Format(es.DbTimeFormat), int(rolltype.RollOut))
	rowCount := sqlDb.Select(traceId, &list, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return nil
	}
	/* *TODO 如果之後有限制回傳筆數
	if rowCount != 1 {
		err = fmt.Errorf("GetRoundCheckList rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return
	}
	*/
	return list
}

// 是否存在rollOut
func IsExistsRolloutHistory(traceId string, gameSequenceNumber string) (hasData bool, rollOutAmount decimal.Decimal) {
	rollOutId := fmt.Sprintf(rollOutIdFormat, gameSequenceNumber)
	sql := `SELECT amount
			FROM RollHistory(nolock) as RO
			WHERE RO.transId=? 
			AND RO.rollType=?`
	params := []interface{}{}
	params = append(params, rollOutId, int(rolltype.RollOut))
	rowCount := sqlDb.Select(traceId, &rollOutAmount, sql, params...)
	//底層錯誤
	if rowCount == -1 {
		return false, decimal.Zero
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRolloutHistory, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "params", params, "rowCount", rowCount)
		return false, decimal.Zero
	}

	//not expected data
	if rollOutAmount.LessThanOrEqual(decimal.Zero) {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRolloutHistory, innererror.TraceNode, traceId, innererror.ErrorInfoNode, dataError, "sql", sql, "params", params, "rollOutAmount", rollOutAmount.String())
		return false, decimal.Zero
	}

	return true, rollOutAmount
}

// 計算連線BetCount
func GetAccountBetCount(traceId string, account string) (count int) {
	sql := `SELECT COUNT(*)
			FROM Account as acct (NOLOCK)
			JOIN GameToken as gt (NOLOCK)
				ON acct.account=gt.account
			JOIN GameResult as gr (NOLOCK)
				ON gt.connectToken=gr.connectToken
			WHERE acct.account=?`
	rowCount := sqlDb.Select(traceId, &count, sql, account)
	//底層錯誤
	if rowCount == -1 {
		return -1
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetAccountBetCount, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "account", account, "rowCount", rowCount)
		return -1
	}

	return count
}

// 計算連線RTP
func GetAccountRtp(traceId string, account string) (rtp int) {
	sql := `SELECT CASE WHEN acct.acctRtp>0 THEN acct.acctRtp
						WHEN ch.chanRtp>0 THEN ch.chanRtp
						WHEN pf.pfRtp>0 THEN pf.pfRtp
					END as RTP
			FROM Account as acct(NOLOCK)
			JOIN Channel as ch(NOLOCK)
			ON acct.chanId=ch.chanId
			JOIN Platform as pf(NOLOCK)
			ON ch.pfCode=pf.pfCode
			WHERE acct.account=?`
	rowCount := sqlDb.Select(traceId, &rtp, sql, account)
	//底層錯誤
	if rowCount == -1 {
		return -1
	}

	//not expected rowCount
	if rowCount != 1 {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetAccountRtp, innererror.TraceNode, traceId, innererror.ErrorInfoNode, rowCountError, "sql", sql, "account", account, "rowCount", rowCount)
		return -1
	}

	return rtp
}
