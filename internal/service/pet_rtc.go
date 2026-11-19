package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	model "project/internal/model"
	"project/pkg/errcode"
	utils "project/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 宠物看护实时音视频（阿里云 ARTC）入会凭证签发。
// AppID / AppKey 只存在于服务端配置（conf.yml 的 pet.rtc 段，可用 GOTP_PET_RTC_* 覆盖），
// 客户端只拿到一次性入会凭证。
//
// 阿里云 ARTC 单参数入会 Token 计算方式（官方）：
//   sign       = hex(sha256(appId + appKey + channelId + userId + nonce + timestamp))
//   base64Token = base64(JSON{appid,channelid,userid,nonce,timestamp,token=sign})
// 客户端以 base64Token 调用 SDK 的 joinChannel(token, userName)。

const (
	petRTCChannelPrefix     = "pet_"
	petRTCDefaultTTLSeconds = 7200
)

type petRTCConfig struct {
	AppID      string
	AppKey     string
	TTLSeconds int64
}

// loadPetRTCConfig 从配置文件读取宠物看护的 RTC 配置
func loadPetRTCConfig() petRTCConfig {
	cfg := petRTCConfig{
		AppID:      viper.GetString("pet.rtc.app_id"),
		AppKey:     viper.GetString("pet.rtc.app_key"),
		TTLSeconds: viper.GetInt64("pet.rtc.token_ttl_seconds"),
	}
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = petRTCDefaultTTLSeconds
	}
	return cfg
}

func (c petRTCConfig) validate() error {
	if c.AppID == "" || c.AppKey == "" {
		return errcode.WithData(errcode.CodeSystemError, map[string]interface{}{
			"rtc": "实时音视频通道未配置：请设置 conf.yml 的 pet.rtc.app_id / pet.rtc.app_key，或环境变量 GOTP_PET_RTC_APP_ID / GOTP_PET_RTC_APP_KEY",
		})
	}
	return nil
}

// GenerateRtcToken 签发指定设备的入会凭证；频道按设备维度固定，同一设备的多位主人进入同一频道
func (p *Pet) GenerateRtcToken(req *model.PetRtcTokenReq, claims *utils.UserClaims) (*model.PetRtcTokenResp, error) {
	if err := p.checkDevice(req.DeviceID, claims.TenantID); err != nil {
		return nil, err
	}

	cfg := loadPetRTCConfig()
	if err := cfg.validate(); err != nil {
		logrus.Error("签发实时音视频凭证失败:", err)
		return nil, err
	}

	channelID := petRTCChannelPrefix + req.DeviceID
	userID := claims.ID
	nonce := ""
	timestamp := time.Now().Unix() + cfg.TTLSeconds

	base64Token, err := buildPetRtcToken(cfg.AppID, cfg.AppKey, channelID, userID, nonce, timestamp)
	if err != nil {
		logrus.Error("生成实时音视频凭证失败:", err)
		return nil, err
	}

	return &model.PetRtcTokenResp{
		AppID:     cfg.AppID,
		ChannelID: channelID,
		UserID:    userID,
		UserName:  claims.Email,
		Token:     base64Token,
		ExpiresAt: timestamp,
	}, nil
}

// buildPetRtcToken 按阿里云 ARTC 单参数入会规范生成 Base64 凭证
func buildPetRtcToken(appID, appKey, channelID, userID, nonce string, timestamp int64) (string, error) {
	signContent := fmt.Sprintf("%s%s%s%s%s%d", appID, appKey, channelID, userID, nonce, timestamp)
	sum := sha256.Sum256([]byte(signContent))
	sign := hex.EncodeToString(sum[:])

	payload := map[string]interface{}{
		"appid":     appID,
		"channelid": channelID,
		"userid":    userID,
		"nonce":     nonce,
		"timestamp": timestamp,
		"token":     sign,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}
