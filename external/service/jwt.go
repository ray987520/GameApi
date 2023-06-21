package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT token資料
type ConnectTokenClaims struct {
	Account  string `json:"account"`
	Currency string `json:"currency"`
	GameId   int    `json:"gameID"`
	jwt.StandardClaims
}

const (
	issuer          = "GameAPI"
	signStringError = "gen token error:%v"
	parseJwtError   = "parse jwt token error:%v"
	getJwtDataError = "get data of jwt error:%v"
)

var (
	jwtSecret []byte // jwt secret key
)

func InitJwt() {
	jwtSecret = []byte(mconfig.GetString("crypt.jwtKey"))
}

// 產生JWT TOKEN
func CreateConnectToken(traceId string, account, currency string, gameId int) (tokenString string) {
	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, esid.JwtValidConnectToken, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("account", account, "currency", currency, "gameId", gameId, "jwtSecret", jwtSecret))
	now := UtcNow()
	claims := new(ConnectTokenClaims)
	claims.Account = account
	claims.Currency = currency
	claims.GameId = gameId
	claims.Issuer = issuer
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(10 * time.Minute).Unix()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		err = fmt.Errorf(signStringError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JwtCreateConnectToken, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage(innererror.ErrorInfoNode, err, "tokenString", tokenString))
		return ""
	}
	return tokenString
}

// 驗證JWT TOKEN
func ValidConnectToken(traceId string, tokenString string) *ConnectTokenClaims {
	zaplog.Infow(innererror.InfoNode, innererror.FunctionNode, esid.JwtValidConnectToken, innererror.TraceNode, traceId, innererror.DataNode, tracer.MergeMessage("tokenString", tokenString, "jwtSecret", jwtSecret))
	//解析jwt token
	token, err := jwt.ParseWithClaims(tokenString, &ConnectTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		err = fmt.Errorf(parseJwtError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JwtValidConnectToken, innererror.TraceNode, traceId, innererror.DataNode, err)
		return nil
	}

	// 從 raw token 中取回資訊
	if claims, ok := token.Claims.(*ConnectTokenClaims); ok && token.Valid {
		return claims
	}

	//無法成功取回對應格式資料就是jwt字串異常
	err = fmt.Errorf(getJwtDataError, err)
	zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JwtValidConnectToken, innererror.TraceNode, traceId, innererror.DataNode, err)
	return nil
}
