package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

type JWT struct {
	Key interface{}
}

// CTokenKeyPrefix C 端令牌在 Redis 中的键前缀，与后台令牌隔离。
const CTokenKeyPrefix = "c_token:"

type UserClaims struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	CreateTime time.Time `json:"create_time"`
	Authority  string    `json:"authority"`
	TenantID   string    `json:"tenant_id"`
	jwt.StandardClaims
}

// CUserClaims C 端用户令牌载荷，与后台 UserClaims 完全隔离（独立密钥与签名）。
type CUserClaims struct {
	UserID       string `json:"user_id"`
	Phone        string `json:"phone"`
	SolutionCode string `json:"solution_code"`
	jwt.StandardClaims
}

func NewJWT(key interface{}) *JWT {
	return &JWT{
		Key: key,
	}
}

// 生成token
func (j *JWT) GenerateToken(claims UserClaims) (string, error) {
	claims.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).Unix()
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 生成token
	return tokenClaims.SignedString(j.Key)
}

// 解析token
func (j *JWT) ParseToken(token string) (*UserClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &UserClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return j.Key, nil
	})
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	if claims, ok := tokenClaims.Claims.(*UserClaims); ok && tokenClaims.Valid {
		return claims, nil
	}
	logrus.Error(err.Error())
	return nil, err
}

// GenerateCToken 生成 C 端用户令牌
func (j *JWT) GenerateCToken(claims CUserClaims) (string, error) {
	claims.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).Unix()
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString(j.Key)
}

// ParseCToken 解析 C 端用户令牌
func (j *JWT) ParseCToken(token string) (*CUserClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &CUserClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return j.Key, nil
	})
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	if claims, ok := tokenClaims.Claims.(*CUserClaims); ok && tokenClaims.Valid {
		return claims, nil
	}
	logrus.Error(err.Error())
	return nil, err
}
