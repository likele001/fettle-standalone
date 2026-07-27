package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

var ErrInvalidKey = errors.New("invalid encryption key")

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	return data[:len(data)-padding], nil
}

func getAESKey() []byte {
	key := os.Getenv("AES_KEY")
	if key == "" {
		key = "fettle_aes_secret_key_32bytes"
	}
	keyBytes := []byte(key)
	if len(keyBytes) < 16 {
		keyBytes = append(keyBytes, make([]byte, 16-len(keyBytes))...)
	} else if len(keyBytes) < 24 {
		keyBytes = append(keyBytes, make([]byte, 24-len(keyBytes))...)
	} else if len(keyBytes) < 32 {
		keyBytes = append(keyBytes, make([]byte, 32-len(keyBytes))...)
	} else {
		keyBytes = keyBytes[:32]
	}
	return keyBytes
}

func AESEncrypt(plaintext string) (string, error) {
	key := getAESKey()
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	paddedData := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)
	
	encrypted := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func AESDecrypt(encryptedText string) (string, error) {
	key := getAESKey()
	
	encrypted, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(encrypted) < aes.BlockSize {
		return "", errors.New("invalid encrypted data")
	}

	iv := encrypted[:aes.BlockSize]
	ciphertext := encrypted[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	paddedData := make([]byte, len(ciphertext))
	mode.CryptBlocks(paddedData, ciphertext)

	data, err := pkcs7Unpad(paddedData)
	if err != nil {
		return "", err
	}

	return string(data), nil
}