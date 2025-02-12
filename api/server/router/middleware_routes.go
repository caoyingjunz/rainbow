package router

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/caoyingjunz/pixiulib/strutil"
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

func NewMiddlewares(o *options.ServerOptions) {
	o.HttpEngine.Use(
		Authentication(o),
	)
}

// Authentication 身份认证
func Authentication(o *options.ServerOptions) gin.HandlerFunc {
	cfg := o.ComponentConfig
	auth := cfg.Server.Auth

	return func(c *gin.Context) {
		if cfg.Default.Mode == "debug" {
			return
		}

		accessKey := c.GetHeader("accessKey")
		if accessKey != auth.AccessKey {
			httputils.AbortFailedWithCode(c, http.StatusUnauthorized, fmt.Errorf("invalid Access Key"))
			return
		}

		timestamp := c.GetHeader("timestamp")
		if err := verifyTimeStamp(timestamp); err != nil {
			httputils.AbortFailedWithCode(c, http.StatusUnauthorized, err)
			return
		}

		signature := c.GetHeader("signature")
		if !verifySignature(accessKey, auth.SecretKey, signature, timestamp) {
			httputils.AbortFailedWithCode(c, http.StatusUnauthorized, fmt.Errorf("invalid Signature"))
			return
		}
	}
}

func verifyTimeStamp(timestamp string) error {
	ts, err := strutil.ParseInt64(timestamp)
	if err != nil {
		return fmt.Errorf("invalid Timestamp %s %v", timestamp, err)
	}
	if time.Now().Unix()-ts > 60*5 {
		return fmt.Errorf("timestamp expired")
	}

	return nil
}

func verifySignature(accessKey, secretKey, signature, timestamp string) bool {
	// 构造签名字符串
	message := fmt.Sprintf("ak=%s&timestamp=%s", accessKey, timestamp)

	// 使用HMAC-SHA256算法生成签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// 比较签名
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
