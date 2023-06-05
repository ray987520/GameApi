package es

import (
	"TestAPI/external/service/mconfig"
	"errors"
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
	issuer = "GameAPI"
)

// jwt secret key
var (
	jwtSecret = []byte(mconfig.GetString("crypt.jwtKey"))
)

// 產生JWT TOKEN
func CreateConnectToken(traceMap string, account, currency string, gameId int) (tokenString string, err error) {
	now := UtcNow()
	claims := new(ConnectTokenClaims)
	claims.Account = account
	claims.Currency = currency
	claims.GameId = gameId
	claims.Issuer = issuer
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(10 * time.Minute).Unix()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = jwtToken.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 驗證JWT TOKEN
func ValidConnectToken(traceMap string, tokenString string) (*ConnectTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ConnectTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, errors.New("invalid")
	}
	// 從 raw token 中取回資訊
	if claims, ok := token.Claims.(*ConnectTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid")
}
