package database

import (
	"TestAPI/entity"
	esid "TestAPI/enum/externalserviceid"
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &errorMessage, sql, params...)
	if err != nil {
		errorMessage = unknowError
		return
	}
	if rowCount != 1 {
		errorMessage = unknowError
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetExternalErrorMessage, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, fmt.Errorf("GetExternalErrorMessage rowCount error"), "sql", sql, "params", params, "rowCount", rowCount)
		return
	}
	if errorMessage == "" {
		errorMessage = unknowError
		return
	}
	return
}

// 取匯率
func GetCurrencyExchangeRate(traceMap string, currency string) (exchangeRate decimal.Decimal, err error) {
	sql := `SELECT [exchangeRate]
			FROM [Currency](nolock)
			WHERE [currency]=?`
	params := []interface{}{currency}
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &exchangeRate, sql, params...)
	if err != nil {
		exchangeRate = decimal.Zero
		return
	}
	if rowCount != 1 {
		exchangeRate = decimal.Zero
		err = fmt.Errorf("GetCurrencyExchangeRate rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params)
		return
	}
	if exchangeRate.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("Unknow Currency")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyExchangeRate, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "exchangeRate", exchangeRate, "rowCount", rowCount)
		exchangeRate = decimal.Zero
		return
	}
	return
}

// 取玩家資訊
func GetPlayerInfo(traceMap string, account, currency string, gameId int) (data entity.AuthConnectTokenResponse, err error) {
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &data, sql, params...)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetPlayerInfo rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerInfo, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return
	}
	data.GameAccount = fmt.Sprintf("%d_%s", data.ChannelID, data.MemberAccount)
	data.BetCount, err = GetAccountBetCount(es.AddTraceMap(traceMap, sqlid.GetAccountBetCount.String()), account)
	if err != nil {
		return
	}
	data.RTP, err = GetAccountRtp(es.AddTraceMap(traceMap, sqlid.GetAccountRtp.String()), account)
	if err != nil {
		return
	}
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
	rowCount, err := sqlDb.Create(es.AddTraceMap(traceMap, string(esid.SqlCreate)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("AddConnectToken rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddConnectToken, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Update(es.AddTraceMap(traceMap, string(esid.SqlSelect)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("UpdateTokenLocation rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.UpdateTokenLocation, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}
	return err == nil
}

// 連線token是否存活
func GetTokenAlive(traceMap string, token string) (alive bool) {
	if GetConnectTokenCache(es.AddTraceMap(traceMap, redisid.GetConnectTokenCache.String()), token) == "1" {
		alive = true
	}

	sql := `SELECT 1 as alive
		FROM [GameToken](nolock)
		WHERE [connectToken]=?
		AND status=?`
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &alive, sql, token, int(tokenstatus.Actived))
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetTokenAlive rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetTokenAlive, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "token", token, "status", int(tokenstatus.Actived), "rowCount", rowCount)
		return false
	}
	if alive {
		SetConnectTokenCache(es.AddTraceMap(traceMap, redisid.SetConnectTokenCache.String()), token, 0)
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
	rowCount, err := sqlDb.Update(es.AddTraceMap(traceMap, string(esid.SqlUpdate)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("DeleteToken rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.DeleteToken, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}
	SetConnectTokenCache(es.AddTraceMap(traceMap, redisid.SetConnectTokenCache.String()), token, 1800)
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
	betTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.BetTime)
	if err != nil {
		return false
	}
	serverTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.ServerTime)
	if err != nil {
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
	rowCount, err := sqlDb.Transaction(es.AddTraceMap(traceMap, string(esid.SqlTransaction)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 3 {
		err = fmt.Errorf("AddGameResultReCountWallet rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResultReCountWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	betTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.BetTime)
	if err != nil {
		return false
	}
	serverTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.ServerTime)
	if err != nil {
		return false
	}
	params = append(params, data.Token, data.GameSequenceNumber, data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout,
		data.CurrencyKindContribution, data.CurrencyKindJackPot, data.SequenceID, data.GameRoom, betTime.Format(es.DbTimeFormat), serverTime.Format(es.DbTimeFormat),
		data.FreeGame, data.TurnTimes, data.BetMode)
	rowCount, err := sqlDb.Create(es.AddTraceMap(traceMap, string(esid.SqlCreate)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("AddGameResult rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameResult, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	data, err = GetPlayerWalletCache(es.AddTraceMap(traceMap, redisid.GetPlayerWalletCache.String()), account, currency)
	if err == nil && data.Currency != "" {
		return
	}
	sql := `SELECT [id] as walletId
					,[amount] as currency
					,[currency] as walletCurrency
				FROM [ManCoin]
				WHERE account=?
				AND currency=?`
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &data, sql, account, currency)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetPlayerWallet rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "account", account, "currency", currency, "rowCount", rowCount)
		return
	}
	if data.Currency == "" {
		err = fmt.Errorf("Get No Wallet")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetPlayerWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data.Currency", data.Currency)
		return
	}
	SetPlayerWalletCache(es.AddTraceMap(traceMap, redisid.SetPlayerWalletCache.String()), account, currency, data)
	return
}

// 以connectToken/將號取GameResult是否存在
func IsExistsTokenGameResult(traceMap string, token, gameSeqNo string) (data bool) {
	sql := `SELECT 1
			FROM [GameResult](nolock)
			WHERE connectToken=?
			AND gameSequenceNumber=?`
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &data, sql, token, gameSeqNo)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("IsExistsTokenGameResult rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsTokenGameResult, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "token", token, "gameSeqNo", gameSeqNo, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &data, sql, token, rollInId)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("IsExistsRollInHistory rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRollInHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "token", token, "rollInId", rollInId, "rowCount", rowCount)
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
	walletId, err := strconv.Atoi(wallet.WalletID)
	if err != nil {
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollInHistory, innererror.ErrorTypeNode, innererror.StringToIntError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "walletId", wallet.WalletID)
		return false
	}
	params = append(params, []interface{}{walletId})
	params = append(params, []interface{}{data.CurrencyKindPayout, walletId})
	rowCount, err := sqlDb.Transaction(es.AddTraceMap(traceMap, string(esid.SqlTransaction)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 3 {
		err = fmt.Errorf("AddRollInHistory rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollInHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	betTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.BetTime)
	if err != nil {
		return false
	}
	params = append(params, data.Token, data.GameSequenceNumber, data.SequenceID, data.GameLog, data.CurrencyKindBet.Mul(exchangeRate), data.CurrencyKindWinLose.Mul(exchangeRate), data.CurrencyKindPayout.Mul(exchangeRate), data.CurrencyKindContribution.Mul(exchangeRate), data.CurrencyKindJackPot.Mul(exchangeRate), data.CurrencyKindBet, data.CurrencyKindWinLose, data.CurrencyKindPayout, data.CurrencyKindContribution, data.CurrencyKindJackPot, betTime.Format(es.DbTimeFormat))
	rowCount, err := sqlDb.Create(es.AddTraceMap(traceMap, string(esid.SqlCreate)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("AddGameLog rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddGameLog, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return false
	}
	return err == nil
}

// 取遊戲語系
func GetGameLanguage(traceMap string, gameId int) (data string, err error) {
	sql := `SELECT lang
			FROM Game(nolock)
			WHERE gameId=?`
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &data, sql, gameId)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetGameLanguage rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetGameLanguage, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "gameId", gameId, "rowCount", rowCount)
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
	rollTime, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, data.RollTime)
	if err != nil {
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
	rowCount, err := sqlDb.Transaction(es.AddTraceMap(traceMap, string(esid.SqlTransaction)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 3 {
		err = fmt.Errorf("AddRollOutHistory rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddRollOutHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Create(es.AddTraceMap(traceMap, string(esid.SqlCreate)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("AddActivityRank rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.AddActivityRank, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &hasData, sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 1 {
		err = fmt.Errorf("IsExistsUnpayActivityDistribution rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsUnpayActivityDistribution, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Transaction(es.AddTraceMap(traceMap, string(esid.SqlTransaction)), sql, params...)
	if err != nil {
		return false
	}
	if rowCount != 3 {
		err = fmt.Errorf("ActivityDistribution rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.ActivityDistribution, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &temp, sql, data.ActivityIV, data.Rank)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetDistributionWallet rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetDistributionWallet, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "ActivityIV", data.ActivityIV, "Rank", data.Rank, "rowCount", rowCount)
		return
	}
	wallet, err = GetPlayerWallet(es.AddTraceMap(traceMap, sqlid.GetPlayerWallet.String()), temp.Account, temp.Currency)
	if err != nil {
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &list, sql)
	if err != nil {
		return
	}
	if rowCount == 0 {
		err = fmt.Errorf("GetCurrencyList rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetCurrencyList, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "rowCount", rowCount)
		return
	}
	return
}

// 查RollOut對應GameResult/RollIn
func GetRoundCheckList(traceMap string, fromDate, toDate string) (list []entity.RoundCheckToken, err error) {
	rollTimeStart, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, fromDate)
	if err != nil {
		return
	}
	rollTimeEnd, err := es.ParseTime(es.AddTraceMap(traceMap, string(esid.ParseTime)), es.ApiTimeFormat, toDate)
	if err != nil {
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
	params := []interface{}{}
	params = append(params, rollTimeStart.Format(es.DbTimeFormat), rollTimeEnd.Format(es.DbTimeFormat), int(rolltype.RollOut))
	_, err = sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &list, sql, params...)
	if err != nil {
		return
	}
	/* TODO 如果之後有限制回傳筆數
	if rowCount != 1 {
		err = fmt.Errorf("GetRoundCheckList rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetRoundCheckList, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return
	}
	*/
	return
}

// 是否存在rollOut
func IsExistsRolloutHistory(traceMap string, gameSequenceNumber string) (hasData bool, rollOutAmount decimal.Decimal) {
	rollOutId := fmt.Sprintf(rollOutIdFormat, gameSequenceNumber)
	sql := `SELECT amount
			FROM RollHistory(nolock) as RO
			WHERE RO.transId=? 
			AND RO.rollType=?`
	params := []interface{}{}
	params = append(params, rollOutId, int(rolltype.RollOut))
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &rollOutAmount, sql, params...)
	if err != nil {
		return false, decimal.Zero
	}
	if rowCount != 1 {
		err = fmt.Errorf("IsExistsRolloutHistory rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRolloutHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rowCount", rowCount)
		return false, decimal.Zero
	}
	if rollOutAmount.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("IsExistsRolloutHistory data error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.IsExistsRolloutHistory, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "params", params, "rollOutAmount", rollOutAmount.String())
		return false, decimal.Zero
	}
	hasData = true
	return
}

// 計算連線BetCount
func GetAccountBetCount(traceMap string, account string) (count int, err error) {
	sql := `SELECT COUNT(*)
			FROM Account as acct (NOLOCK)
			JOIN GameToken as gt (NOLOCK)
				ON acct.account=gt.account
			JOIN GameResult as gr (NOLOCK)
				ON gt.connectToken=gr.connectToken
			WHERE acct.account=?`
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &count, sql, account)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetAccountBetCount rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetAccountBetCount, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "account", account, "rowCount", rowCount)
		return
	}
	return
}

// 計算連線RTP
func GetAccountRtp(traceMap string, account string) (rtp int, err error) {
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
	rowCount, err := sqlDb.Select(es.AddTraceMap(traceMap, string(esid.SqlSelect)), &rtp, sql, account)
	if err != nil {
		return
	}
	if rowCount != 1 {
		err = fmt.Errorf("GetAccountRtp rowCount error")
		zaplog.Errorw(innererror.DBSqlError, innererror.FunctionNode, sqlid.GetAccountRtp, innererror.ErrorTypeNode, innererror.SelectError, innererror.ErrorInfoNode, err, "sql", sql, "account", account, "rowCount", rowCount)
		return
	}
	return
}
