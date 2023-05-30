package es

import (
	"TestAPI/external/service/mconfig"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	chiperError     = "could not create new cipher: %v"
	encryptError    = "could not encrypt: %v"
	base64Error     = "could not base64 decode: %v"
	bloackSizeError = "invalid ciphertext block size"
)

// AES128加密金鑰,固定16Bytes
var aes128Key = mconfig.GetString("crypt.aes128Key")

// AES128加密,輸出加密後base64編碼字串
func Aes128Encrypt(traceMap string, message []byte) (string, error) {
	byteMsg := []byte(message)
	block, err := aes.NewCipher([]byte(aes128Key))
	if err != nil {
		return "", fmt.Errorf(chiperError, err)
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	//動態IV,避免同樣資料產生相同加密字串
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf(encryptError, err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// AES128解密,輸入應為加密後base64編碼字串,輸出原資料[]byte
func Aes128Decrypt(traceMap string, message string) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return nil, fmt.Errorf(base64Error, err)
	}

	block, err := aes.NewCipher([]byte(aes128Key))
	if err != nil {
		return nil, fmt.Errorf(chiperError, err)
	}

	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf(bloackSizeError)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
