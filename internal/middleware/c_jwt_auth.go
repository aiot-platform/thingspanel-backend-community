package middleware

import (
	"context"
	"net/http"
	"time"

	"project/pkg/global"
	utils "project/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CAuth C 端用户鉴权中间件：读取 x-token，校验 Redis 与独立密钥签发的令牌，
// 校验通过后把 *utils.CUserClaims 写入上下文键 "c_claims"。
func CAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("x-token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    ErrCodeNoAuth,
				Message: "missing authentication (x-token required)",
			})
			c.Abort()
			return
		}

		if global.REDIS.Get(context.Background(), utils.CTokenKeyPrefix+token).Val() != "1" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    ErrCodeTokenExpired,
				Message: "token has expired",
			})
			c.Abort()
			return
		}

		// 刷新令牌过期时间
		timeout := viper.GetInt("session.timeout")
		if timeout <= 0 {
			timeout = 60
		}
		global.REDIS.Set(context.Background(), utils.CTokenKeyPrefix+token, "1", time.Duration(timeout)*time.Minute)

		key := viper.GetString("jwt.c_secret")
		if key == "" {
			key = viper.GetString("jwt.key")
		}
		j := utils.NewJWT([]byte(key))
		claims, err := j.ParseCToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    ErrCodeInvalidToken,
				Message: "invalid token format",
			})
			c.Abort()
			return
		}

		c.Set("c_claims", claims)
		c.Next()
	}
}

// SaveCToken 将 C 端令牌写入 Redis。
func SaveCToken(token string, ttl time.Duration) error {
	if global.REDIS == nil {
		logrus.Error("redis client is not initialized, c-end token not saved")
		return nil
	}
	return global.REDIS.Set(context.Background(), utils.CTokenKeyPrefix+token, "1", ttl).Err()
}

// DeleteCToken 删除 C 端令牌。
func DeleteCToken(token string) error {
	if global.REDIS == nil {
		return nil
	}
	return global.REDIS.Del(context.Background(), utils.CTokenKeyPrefix+token).Err()
}
