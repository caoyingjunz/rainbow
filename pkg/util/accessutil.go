package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// GenerateAK 生成 Access Key (含前缀)
func GenerateAK(prefix string) (string, error) {
	b := make([]byte, 12) // 12 字节随机数，hex 编码后 24 字符
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

// GenerateSK 生成 Secret Key (Base64)
func GenerateSK() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
