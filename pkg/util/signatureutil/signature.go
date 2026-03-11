package signatureutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
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
	accessKay := c.GetHeader("X-ACCESS-KEY")
	if len(accessKay) == 0 {
		return fmt.Errorf("missing AccessKay")
	}

	obj, err := f.Access().Get(c, accessKay)
	if err != nil {
		klog.Errorf("获取 ak(%s) sk 记录失败 %v", accessKay, err)
		return fmt.Errorf("invalid AccessKay")
	}
	// 判断ak sk是否已过期或者关闭
	if obj.ExpireTime != nil && obj.ExpireTime.Before(time.Now()) {
		return fmt.Errorf("expireTime AccessKay")
	}

	expected := GenerateSignature(
		map[string]string{"action": "pullOrCacheRepo", "accessKey": accessKay},
		[]byte(obj.SecretKey))

	if hmac.Equal([]byte(expected), []byte(c.GetHeader("Authorization"))) {
		return nil
	}

	return fmt.Errorf("invaild Signature")
}
