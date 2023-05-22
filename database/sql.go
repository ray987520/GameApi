package database

import (
	"TestAPI/entity"
	"TestAPI/enum/innererror"
	"TestAPI/enum/rankstatus"
	"TestAPI/enum/redisid"
	"TestAPI/enum/rolltype"
	"TestAPI/enum/sqlid"
	"TestAPI/enum/tokenlocation"
	"TestAPI/enum/tokenstatus"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

const (
	rollOutIdFormat = "rollOut-%s"   //rollOut transactionId Format
	rollInIdFormat  = "rollIn-%s"    //rollIn transactionId Format
	unknowError     = "Unknow Error" //unknow error default message
)

var sqlDb iface.ISqlService

// 注入sql client
func InitSqlWorker(db iface.ISqlService) bool {
	sqlDb = db
	return true
}

// 取對外輸出的錯誤訊息
func GetExternalErrorMessage(traceMap string, code string) (errorMessage string) {
	sql := `SELECT message
	FROM [dbo].[ErrorMessage](nolock)
	WHERE code=?
	AND codeType=1
	AND langCode='en-US'`
	params := []interface{}{code}
	err := sqlDb.Select(&errorMessage, sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetExternalErrorMessage, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "code", code)
	}
	if errorMessage == "" {
		errorMessage = unknowError
	}
	return
}

// 取匯率
func GetCurrencyExchangeRate(traceMap string, currency string) (exchangeRate decimal.Decimal, err error) {
	sql := `SELECT [exchangeRate]
			FROM [Currency](nolock)
			WHERE [currency]=?`
	params := []interface{}{currency}
	err = sqlDb.Select(&exchangeRate, sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "currency", currency)
		return
	}
	if exchangeRate.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("Unknow Currency")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "exchangeRate", exchangeRate)
		return
	}
	return
}

// 取玩家資訊
func GetPlayerInfo(traceMap string, account, currency string, gameId int) (data entity.AuthConnectTokenResponse) {
	sql := `SELECT ch.pfCode as platformID
					,ch.chanId as channelID
					,'' as poolID
					,gc.gameId as gameID
					,acct.acctId as memberID
					,gc.rtp as RTP
					,'' as gameAccount
					,acct.account as memberAccount
					,mc.id as walletId
					,mc.currency as currencyKind
					,mc.currency as walletCurrency
					,mc.amount as currency
					,gc.threshold
					,gc.app
					,gc.report
					,gc.gamePlat
					,0 as betCount
				FROM Account as acct(nolock)
				JOIN Channel as ch(nolock)
					ON acct.chanId=ch.chanId
				JOIN ManCoin as mc(nolock)
					ON acct.account=mc.account
				JOIN GameConfig as gc(nolock)
					ON gc.chanId=ch.chanId AND gc.gameId=?
				WHERE acct.account=?
					AND mc.currency=?`
	params := []interface{}{gameId, account, currency}
	err := sqlDb.Select(&data, sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerInfo, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "gameId", gameId, "account", account, "currency", currency)
		return
	}
	data.GameAccount = fmt.Sprintf("%d_%s", data.ChannelID, data.MemberAccount)
	return
}

// 建立連線token
func AddConnectToken(traceMap string, token, account, currency, ip string, gameId int, now time.Time) bool {
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
	err := sqlDb.Create(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerInfo, innererror.ErrorTypeNode, innererror.CreateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "gameId", gameId, "account", account, "currency", currency, "now", now.Format(es.DbTimeFormat), "ip", ip, "location", int(tokenlocation.Default), "status", int(tokenstatus.Actived))
		return false
	}
	return err == nil
}

// 刷新連線token位置
func UpdateTokenLocation(traceMap string, token string, location int) bool {
	sql := `UPDATE [dbo].[GameToken]
			SET [location] = ?
 			WHERE [connectToken]=?`
	params := []interface{}{location, token}
	err := sqlDb.Update(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerInfo, innererror.ErrorTypeNode, innererror.UpdateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "location", location)
		return false
	}
	return err == nil
}

// 連線token是否存活
func GetTokenAlive(traceMap string, token string) (alive bool) {
	if GetConnectTokenCache(es.AddTraceMap(traceMap, redisid.GetConnectTokenCache.String()), token) == "1" {
		alive = true
	} else {
		sql := `SELECT 1 as alive
		FROM [GameToken](nolock)
		WHERE [connectToken]=?
		AND status=?`
		err := sqlDb.Select(&alive, sql, token, int(tokenstatus.Actived))
		if err != nil {
			zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetTokenAlive, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "status", int(tokenstatus.Actived))
			return false
		}
		if alive {
			SetConnectTokenCache(es.AddTraceMap(traceMap, redisid.SetConnectTokenCache.String()), token, 0)
		}
	}

	return
}

// 登出刪除連線token
func DeleteToken(traceMap string, token string, deleteTime time.Time) bool {
	sql := `UPDATE [dbo].[GameToken]
			SET [status] = ?
				,deleteTime=?
 			WHERE [connectToken]=?`
	params := []interface{}{int(tokenstatus.Deleted), deleteTime.Format(es.DbTimeFormat), token}
	err := sqlDb.Update(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.DeleteToken, innererror.ErrorTypeNode, innererror.UpdateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "status", int(tokenstatus.Deleted), "deleteTime", deleteTime.Format(es.DbTimeFormat))
		return false
	} else {
		SetConnectTokenCache(es.AddTraceMap(traceMap, redisid.SetConnectTokenCache.String()), token, 1800)
	}
	return err == nil
}

// 添加GameResult|RollHistory並更新錢包
func AddGameResultReCountWallet(traceMap string, data entity.GameResult, wallet entity.PlayerWallet, now time.Time) bool {
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
	betTime, err := time.Parse(es.ApiTimeFormat, data.BetTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "betTime", data.BetTime)
		return false
	}
	serverTime, err := time.Parse(es.ApiTimeFormat, data.ServerTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "serverTime", data.ServerTime)
		return false
	}
	params = append(params, []interface{}{data.Token, data.GameSequenceNumber, data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout,
		data.CurrencyKindContribution, data.CurrencyKindJackPot, data.SequenceID, data.GameRoom, betTime.Format(es.DbTimeFormat), serverTime.Format(es.DbTimeFormat),
		data.FreeGame, data.TurnTimes, data.BetMode})
	walletId, err := strconv.Atoi(wallet.WalletID)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.StringToIntError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletId", wallet.WalletID)
		return false
	}
	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.CurrencyKindPayout.Sub(data.CurrencyKindBet), walletId})
	err = sqlDb.Transaction(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.TransactionError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data, "wallet", wallet, "now", now.Format(es.DbTimeFormat))
		return false
	}
	return err == nil
}

// 添加賽果
func AddGameResult(traceMap string, data entity.GameResult) bool {
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
	betTime, err := time.Parse(es.ApiTimeFormat, data.BetTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResult, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "betTime", data.BetTime)
		return false
	}
	serverTime, err := time.Parse(es.ApiTimeFormat, data.ServerTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResult, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "serverTime", data.ServerTime)
		return false
	}
	params = append(params, data.Token, data.GameSequenceNumber, data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout,
		data.CurrencyKindContribution, data.CurrencyKindJackPot, data.SequenceID, data.GameRoom, betTime.Format(es.DbTimeFormat), serverTime.Format(es.DbTimeFormat),
		data.FreeGame, data.TurnTimes, data.BetMode)
	err = sqlDb.Create(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.CreateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return false
	}
	return err == nil
}

// 補單token是否存活
func GetFinishGameResultTokenAlive(traceMap string, token string) (alive bool) {
	if GetFinishGameResultTokenCache(es.AddTraceMap(traceMap, redisid.GetFinishGameResultTokenCache.String()), token) == "1" {
		alive = true
	}
	return
}

// 取玩家錢包
func GetPlayerWallet(traceMap string, account, currency string) (data entity.PlayerWallet, err error) {
	if data, err = GetPlayerWalletCache(es.AddTraceMap(traceMap, redisid.GetPlayerWalletCache.String()), account, currency); err == nil && data.Currency != "" {
		return
	} else {
		sql := `SELECT [id] as walletId
					,[amount] as currency
					,[currency] as walletCurrency
				FROM [ManCoin]
				WHERE account=?
				AND currency=?`
		err = sqlDb.Select(&data, sql, account, currency)
		if err != nil {
			zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "account", account, "currency", currency)
			return
		} else {
			SetPlayerWalletCache(es.AddTraceMap(traceMap, redisid.SetPlayerWalletCache.String()), account, currency, data)
		}
	}
	return
}

// 以connectToken/將號取GameResult是否存在
func IsExistsTokenGameResult(traceMap string, token, gameSeqNo string) (data bool) {
	sql := `SELECT 1
			FROM [GameResult](nolock)
			WHERE connectToken=?
			AND gameSequenceNumber=?`
	err := sqlDb.Select(&data, sql, token, gameSeqNo)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsTokenGameResult, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "gameSeqNo", gameSeqNo)
		return false
	}
	return
}

// 以connectToken/將號取RollInHistory是否存在
func IsExistsRollInHistory(traceMap string, token, gameSeqNo string) (data bool) {
	rollInId := fmt.Sprintf(rollInIdFormat, gameSeqNo)
	sql := `SELECT 1
			FROM RollHistory(nolock)
			WHERE connectToken=?
			AND transId=?`
	err := sqlDb.Select(&data, sql, token, rollInId)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRollInHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "token", token, "rollInId", rollInId)
		return false
	}
	return
}

// RollIn並更新錢包
func AddRollInHistory(traceMap string, data entity.GameResult, wallet entity.PlayerWallet, now time.Time) bool {
	sql := []string{`INSERT INTO [dbo].[RollHistory]
           ([connectToken]
           ,[transId]
           ,[gameSequenceNumber]
           ,[amount]
           ,[currnecy]
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
	walletId, err := strconv.Atoi(wallet.WalletID)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollInHistory, innererror.ErrorTypeNode, innererror.StringToIntError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletId", wallet.WalletID)
		return false
	}
	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.CurrencyKindPayout, walletId})
	err = sqlDb.Transaction(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollInHistory, innererror.ErrorTypeNode, innererror.TransactionError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data, "wallet", wallet, "now", now.Format(es.DbTimeFormat))
		return false
	}
	return err == nil
}

// 添加遊戲紀錄
func AddGameLog(traceMap string, data entity.GameLog, exchangeRate decimal.Decimal) bool {
	sql := `INSERT INTO [dbo].[GameLog]
				([connectToken]
				,[gameSequenceNumber]
				,[sequenceID]
				,[gameLog]
				,[bet]
				,[winLose]
				,[payout]
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
	betTime, err := time.Parse(es.ApiTimeFormat, data.BetTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameLog, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "betTime", data.BetTime)
		return false
	}
	params = append(params, data.Token, data.GameSequenceNumber, data.SequenceID, data.GameLog, data.CurrencyKindBet.Mul(exchangeRate), data.CurrencyKindWinLose.Mul(exchangeRate), data.CurrencyKindPayout.Mul(exchangeRate), data.CurrencyKindContribution.Mul(exchangeRate), data.CurrencyKindJackPot.Mul(exchangeRate), data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout, data.CurrencyKindContribution, data.CurrencyKindJackPot, betTime.Format(es.DbTimeFormat))
	err = sqlDb.Create(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameLog, innererror.ErrorTypeNode, innererror.CreateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data, "exchangeRate", exchangeRate)
		return false
	}
	return err == nil
}

// 取遊戲語系
func GetGameLanguage(traceMap string, gameId int) (data string, err error) {
	sql := `SELECT lang
			FROM Game(nolock)
			WHERE gameId=?`
	err = sqlDb.Select(&data, sql, gameId)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetGameLanguage, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "gameId", gameId)
		return
	}
	return
}

// RollOut並更新錢包
func AddRollOutHistory(traceMap string, data entity.RollHistory, wallet entity.PlayerWallet) bool {
	sql := []string{`INSERT INTO [dbo].[RollHistory]
           ([connectToken]
           ,[transId]
           ,[gameSequenceNumber]
           ,[amount]
           ,[currnecy]
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
	rollTime, err := time.Parse(es.ApiTimeFormat, data.RollTime)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollOutHistory, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "rollTime", data.RollTime)
		return false
	}
	params = append(params, []interface{}{data.Token, data.TransID, data.GameSequenceNumber, data.Amount, wallet.Currency, int(rolltype.RollOut), rollTime.Format(es.DbTimeFormat)})
	walletId, err := strconv.Atoi(wallet.WalletID)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollOutHistory, innererror.ErrorTypeNode, innererror.StringToIntError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletId", wallet.WalletID)
		return false
	}
	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.Amount.Neg(), walletId})
	err = sqlDb.Transaction(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollOutHistory, innererror.ErrorTypeNode, innererror.TransactionError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data, "wallet", wallet)
		return false
	}
	return err == nil
}

// 添加活動未派彩紀錄
func AddActivityRank(traceMap string, data entity.Settlement) bool {
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
	err := sqlDb.Create(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddActivityRank, innererror.ErrorTypeNode, innererror.CreateError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return false
	}
	return err == nil
}

// 取是否有對應未派彩紀錄
func IsExistsUnpayActivityDistribution(traceMap string, activityIV string, rank int) (hasData bool) {
	sql := `SELECT 1
			FROM [dbo].[ActivityRank]
			WHERE [activityIV]=?
			AND [rank]=?
			AND status=?`
	params := []interface{}{}
	params = append(params, activityIV, rank, int(rankstatus.UnPay))
	err := sqlDb.Select(&hasData, sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsUnpayActivityDistribution, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "activityIV", activityIV, "rank", rank, "status", int(rankstatus.UnPay))
		return false
	}
	return
}

// 派彩並更新錢包
func ActivityDistribution(traceMap string, data entity.Distribution, walletID string, now time.Time) bool {
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
	walletId, err := strconv.Atoi(walletID)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.ActivityDistribution, innererror.ErrorTypeNode, innererror.StringToIntError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletID", walletID)
		return false
	}
	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.PrizePayout, walletId})
	err = sqlDb.Transaction(sql, params...)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.ActivityDistribution, innererror.ErrorTypeNode, innererror.TransactionError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data, "walletID", walletID, "now", now.Format(es.DbTimeFormat))
		return false
	}
	return err == nil
}

// 取派彩對象玩家錢包
func GetDistributionWallet(traceMap string, data entity.Distribution) (account string, wallet entity.PlayerWallet, err error) {
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
	err = sqlDb.Select(&temp, sql, data.ActivityIV, data.Rank)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetDistributionWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
		return
	}
	wallet, err = GetPlayerWallet(es.AddTraceMap(traceMap, sqlid.GetPlayerWallet.String()), temp.Account, temp.Currency)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetDistributionWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "temp", temp)
		return
	}
	return temp.Account, wallet, err
}

// 取支援的Currency清單
func GetCurrencyList(traceMap string) (list []entity.CurrencyListResponse, err error) {
	sql := `SELECT [id]
				,[currency]
				,[exchangeRate]
			FROM [Currency]`
	err = sqlDb.Select(&list, sql)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyList, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
		return
	}
	return
}

// 查RollOut對應GameResult/RollIn
func GetRoundCheckList(traceMap string, fromDate, toDate string) (list []entity.RoundCheckToken, err error) {
	rollTimeStart, err := time.Parse(es.ApiTimeFormat, fromDate)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "fromDate", fromDate)
		return
	}
	rollTimeEnd, err := time.Parse(es.ApiTimeFormat, toDate)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.TimeParseError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "format", es.ApiTimeFormat, "toDate", toDate)
		return
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
	err = sqlDb.Select(&list, sql, rollTimeStart.Format(es.DbTimeFormat), rollTimeEnd.Format(es.DbTimeFormat), int(rolltype.RollOut))
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "rollTimeStart", rollTimeStart.Format(es.DbTimeFormat), "rollTimeEnd", rollTimeEnd.Format(es.DbTimeFormat), "rollType", int(rolltype.RollOut))
		return
	}
	return
}

// 是否存在rollOut
func IsExistsRolloutHistory(traceMap string, gameSequenceNumber string) (hasData bool) {
	rollOutId := fmt.Sprintf(rollOutIdFormat, gameSequenceNumber)
	sql := `SELECT 1
			FROM RollHistory(nolock) as RO
			WHERE RO.transId=? 
			AND RO.rollType=?`
	err := sqlDb.Select(&hasData, sql, rollOutId, int(rolltype.RollOut))
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "rollOutId", rollOutId, "rollType", int(rolltype.RollOut))
		return false
	}
	return
}
