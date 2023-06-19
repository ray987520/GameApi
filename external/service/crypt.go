package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/mconfig"
	"TestAPI/external/service/zaplog"
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

var aes128Key string // AES128加密金鑰,固定16Bytes

func InitCrypt() {
	aes128Key = mconfig.GetString("crypt.aes128Key")
}

// AES128加密,輸出加密後base64編碼字串
func Aes128Encrypt(traceId string, message []byte) string {
	byteMsg := []byte(message)
	block, err := aes.NewCipher([]byte(aes128Key))
	if err != nil {
		err = fmt.Errorf(chiperError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.Aes128Encrypt, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return ""
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	//動態IV,避免同樣資料產生相同加密字串
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		err = fmt.Errorf(encryptError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.Aes128Encrypt, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText)
}

// AES128解密,輸入應為加密後base64編碼字串,輸出原資料[]byte
func Aes128Decrypt(traceId string, message string) []byte {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		err = fmt.Errorf(base64Error, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.Aes128Decrypt, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return nil
	}

	block, err := aes.NewCipher([]byte(aes128Key))
	if err != nil {
		err = fmt.Errorf(chiperError, err)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.Aes128Decrypt, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return nil
	}

	if len(cipherText) < aes.BlockSize {
		err = fmt.Errorf(bloackSizeError)
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.Aes128Decrypt, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return nil
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText
}
