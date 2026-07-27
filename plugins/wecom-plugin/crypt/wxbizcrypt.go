package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

type EncryptRequestXML struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	AgentID    string   `xml:"AgentID"`
	Encrypt    string   `xml:"Encrypt"`
}

type EncryptResponseXML struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string   `xml:"Encrypt"`
	MsgSignature string   `xml:"MsgSignature"`
	TimeStamp    string   `xml:"TimeStamp"`
	Nonce        string   `xml:"Nonce"`
}

type WXBizMsgCrypt struct {
	token          string
	encodingAESKey string
	corpID         string
	aesKey         []byte
}

func NewWXBizMsgCrypt(token, encodingAESKey, corpID string) *WXBizMsgCrypt {
	key, _ := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	return &WXBizMsgCrypt{
		token:          token,
		encodingAESKey: encodingAESKey,
		corpID:         corpID,
		aesKey:         key,
	}
}

func (c *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echostr string) (string, error) {
	sign := c.signature(timestamp, nonce, echostr)
	if sign != msgSignature {
		return "", fmt.Errorf("signature mismatch")
	}
	plaintext, err := c.decrypt(echostr)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

func (c *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce, postData string) (string, error) {
	var encReq EncryptRequestXML
	if err := xml.Unmarshal([]byte(postData), &encReq); err != nil {
		return "", fmt.Errorf("xml unmarshal failed: %w", err)
	}

	sign := c.signature(timestamp, nonce, encReq.Encrypt)
	if sign != msgSignature {
		return "", fmt.Errorf("signature mismatch")
	}

	plaintext, err := c.decrypt(encReq.Encrypt)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

func (c *WXBizMsgCrypt) EncryptMsg(replyMsg, timestamp, nonce string) (string, error) {
	encrypted, err := c.encrypt(replyMsg)
	if err != nil {
		return "", err
	}

	sign := c.signature(timestamp, nonce, encrypted)

	respXML := EncryptResponseXML{
		Encrypt:      encrypted,
		MsgSignature: sign,
		TimeStamp:    timestamp,
		Nonce:        nonce,
	}
	data, err := xml.Marshal(respXML)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *WXBizMsgCrypt) signature(timestamp, nonce, encrypt string) string {
	arr := []string{c.token, timestamp, nonce, encrypt}
	sort.Strings(arr)
	str := strings.Join(arr, "")
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *WXBizMsgCrypt) decrypt(text string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("aes new cipher failed: %w", err)
	}

	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := c.aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	padding := int(decrypted[len(decrypted)-1])
	if padding > 32 || padding > len(decrypted) {
		return "", fmt.Errorf("invalid padding")
	}
	decrypted = decrypted[:len(decrypted)-padding]

	msgLen := binary.BigEndian.Uint32(decrypted[16:20])
	if int(msgLen)+20 > len(decrypted) {
		return "", fmt.Errorf("invalid message length")
	}
	msgContent := decrypted[20 : 20+msgLen]
	corpID := string(decrypted[20+msgLen:])

	if corpID != c.corpID {
		return "", fmt.Errorf("corpid mismatch")
	}

	return string(msgContent), nil
}

func (c *WXBizMsgCrypt) encrypt(text string) (string, error) {
	random := make([]byte, 16)
	binary.BigEndian.PutUint32(random[12:16], uint32(len(text)))
	plaintext := append(random, []byte(text)...)
	plaintext = append(plaintext, []byte(c.corpID)...)

	padding := 32 - len(plaintext)%32
	if padding == 0 {
		padding = 32
	}
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}

	iv := c.aesKey[:aes.BlockSize]
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
