package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

func GetDecryptKey(password string) []byte {
	// spell-checker: disable-next-line
	return pbkdf2.Key([]byte(password), []byte("saltysalt"), 1003, 16, sha1.New)
}

func DecryptChromeCookieValue(encryptedValue []byte, key []byte, version int) (string, error) {
	iv := bytes.Repeat([]byte{0x20}, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(encryptedValue) < aes.BlockSize {
		return "", errors.New("cipher text too short")
	}

	decrypted := make([]byte, len(encryptedValue)-3)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, encryptedValue[3:])

	var cookieText string
	if version >= 24 {
		cookieText = string(decrypted[32:])
	} else {
		cookieText = string(decrypted)
	}

	cookieText = string(unpadding([]byte(cookieText)))

	return cookieText, nil
}

func unpadding(src []byte) []byte {
	length := len(src)
	unpad := int(src[length-1])
	return src[:(length - unpad)]
}
