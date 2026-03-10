package signatureutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/gin-gonic/gin"
	"sort"
	"strings"
)

func GenerateSignature(params map[string]string, secret []byte) string {
	// 1. 提取所有键并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接成 key=value 对，用 & 连接
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}
	message := builder.String()

	// 3. 计算 HMAC-SHA256
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySignature(c *gin.Context, f db.ShareDaoFactory) error {
	//// 1. 提取所有键并排序
	//keys := make([]string, 0, len(params))
	//for k := range params {
	//	keys = append(keys, k)
	//}
	//sort.Strings(keys)
	//
	//// 2. 拼接成 key=value 对，用 & 连接
	//var builder strings.Builder
	//for i, k := range keys {
	//	if i > 0 {
	//		builder.WriteString("&")
	//	}
	//	builder.WriteString(k)
	//	builder.WriteString("=")
	//	builder.WriteString(params[k])
	//}
	//message := builder.String()
	//
	//// 3. 计算 HMAC-SHA256
	//h := hmac.New(sha256.New, secret)
	//h.Write([]byte(message))
	//return hex.EncodeToString(h.Sum(nil))
	return nil
}
