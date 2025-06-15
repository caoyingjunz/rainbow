package router

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"time"

	"github.com/caoyingjunz/pixiulib/httputils"
	"github.com/caoyingjunz/pixiulib/strutil"
	"github.com/gin-gonic/gin"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

func NewMiddlewares(o *options.ServerOptions) {
	o.HttpEngine.Use(
		Authentication(o),
		Limiter(o),
	)
}

// Limiter 限速
func Limiter(o *options.ServerOptions) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(o.ComponentConfig.RateLimit.NormalRateLimit.MaxRequests), o.ComponentConfig.RateLimit.NormalRateLimit.MaxRequests)
	specialLimiter := rate.NewLimiter(rate.Limit(o.ComponentConfig.RateLimit.SpecialRateLimit.MaxRequests), o.ComponentConfig.RateLimit.SpecialRateLimit.MaxRequests)
	return func(c *gin.Context) {
		if isRateLimitedPath(c.Request.URL.Path, o.ComponentConfig.RateLimit.SpecialRateLimit.RateLimitedPath) {
			if !specialLimiter.Allow() {
				httputils.AbortFailedWithCode(c, http.StatusForbidden, fmt.Errorf("too many requests"))
			}
		} else {
			if !limiter.Allow() {
				httputils.AbortFailedWithCode(c, http.StatusForbidden, fmt.Errorf("too many requests"))
			}
		}
	}
}

// 检查请求路径是否在限速列表中
func isRateLimitedPath(path string, rateLimitedPaths []string) bool {
	for _, limitedPath := range rateLimitedPaths {
		if strings.Contains(path, limitedPath) {
			return true
		}
	}
	return false
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
