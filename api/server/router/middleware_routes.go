package router

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/caoyingjunz/pixiulib/httputils"
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
	return func(c *gin.Context) {
		if cfg.Default.Mode == "debug" {
			return
		}
		auth := cfg.Server.Auth

		accessKey := c.GetHeader("accessKey")
		if accessKey != auth.AccessKey {
			httputils.AbortFailedWithCode(c, http.StatusUnauthorized, fmt.Errorf("invalid Access Key"))
			return
		}

		timestamp := c.GetHeader("timestamp")
		signature := c.GetHeader("signature")
		if !verifySignature(accessKey, auth.SecretKey, signature, timestamp) {
			httputils.AbortFailedWithCode(c, http.StatusUnauthorized, fmt.Errorf("invalid Signature"))
			return
		}
	}
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
